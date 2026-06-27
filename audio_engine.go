package pigo8

import (
	"bytes"
	"io"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type audioFormat uint8

const (
	audioFormatPCM audioFormat = iota
	audioFormatF32
)

var (
	audioContextInstance *audio.Context
	audioContextOnce     sync.Once

	audioPlayerOnce    sync.Once
	audioPlayerF32Once sync.Once
	musicEngines       engineSet
)

type engineSet struct {
	mu  sync.RWMutex
	pcm *musicEngine
	f32 *musicEngine
}

type preparedPlayback struct {
	stream io.ReadSeeker
	length int64
}

type musicEngine struct {
	format         audioFormat
	audioContext   *audio.Context
	musicPlayers   map[int]*audio.Player
	musicLoopState map[int]bool
	mutex          sync.Mutex
}

func getAudioContext() *audio.Context {
	audioContextOnce.Do(func() {
		audioContextInstance = audio.NewContext(sampleRate)
	})
	return audioContextInstance
}

func getAudioPlayer() *musicEngine {
	audioPlayerOnce.Do(func() {
		ensureSharedMusicDataLoaded()
		musicEngines.set(audioFormatPCM, newMusicEngine(audioFormatPCM))
	})
	return musicEngines.get(audioFormatPCM)
}

func getAudioPlayerF32() *musicEngine {
	audioPlayerF32Once.Do(func() {
		ensureSharedMusicDataLoaded()
		musicEngines.set(audioFormatF32, newMusicEngine(audioFormatF32))
	})
	return musicEngines.get(audioFormatF32)
}

func (es *engineSet) set(format audioFormat, engine *musicEngine) {
	es.mu.Lock()
	defer es.mu.Unlock()
	switch format {
	case audioFormatPCM:
		es.pcm = engine
	case audioFormatF32:
		es.f32 = engine
	}
}

func (es *engineSet) get(format audioFormat) *musicEngine {
	es.mu.RLock()
	defer es.mu.RUnlock()
	if format == audioFormatF32 {
		return es.f32
	}
	return es.pcm
}

func (es *engineSet) list() []*musicEngine {
	es.mu.RLock()
	defer es.mu.RUnlock()

	engines := make([]*musicEngine, 0, 2)
	if es.pcm != nil {
		engines = append(engines, es.pcm)
	}
	if es.f32 != nil {
		engines = append(engines, es.f32)
	}
	return engines
}

func newMusicEngine(format audioFormat) *musicEngine {
	return &musicEngine{
		format:         format,
		audioContext:   getAudioContext(),
		musicPlayers:   make(map[int]*audio.Player),
		musicLoopState: make(map[int]bool),
	}
}

func newPreparedPlayback(stream io.ReadSeeker) (preparedPlayback, error) {
	length, err := playbackStreamLength(stream)
	if err != nil {
		return preparedPlayback{}, err
	}
	return preparedPlayback{
		stream: stream,
		length: length,
	}, nil
}

func playbackStreamLength(stream io.ReadSeeker) (int64, error) {
	length, err := stream.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if _, err := stream.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	return length, nil
}

func prepareF32Playback(audioData []byte) (preparedPlayback, error) {
	reader := bytes.NewReader(audioData)
	wavReader, err := wav.DecodeF32(reader)
	if err != nil {
		return preparedPlayback{}, err
	}

	var stream io.ReadSeeker = wavReader
	if wavReader.SampleRate() != sampleRate {
		stream = audio.ResampleF32(wavReader, wavReader.Length(), wavReader.SampleRate(), sampleRate)
	}

	return newPreparedPlayback(stream)
}

func (me *musicEngine) reloadAudioFiles() {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	me.stopAllLocked()
}

func (me *musicEngine) pauseAndRewindAllLocked() {
	for _, player := range me.musicPlayers {
		if player != nil {
			player.Pause()
			if err := player.Rewind(); err != nil {
				log.Printf("Error rewinding player: %v", err)
			}
		}
	}
}

func (me *musicEngine) stopAllLocked() {
	for _, player := range me.musicPlayers {
		if player != nil {
			player.Pause()
			if err := player.Close(); err != nil {
				log.Printf("Error closing player: %v", err)
			}
		}
	}
	me.musicPlayers = make(map[int]*audio.Player)
	me.musicLoopState = make(map[int]bool)
}

func (me *musicEngine) play(id int, opts MusicOptions) {
	audioData, exists := getSharedMusicData(id)
	if !exists {
		log.Printf("Warning: Audio file with ID %d not found", id)
		return
	}

	me.mutex.Lock()
	defer me.mutex.Unlock()

	player, exists := me.musicPlayers[id]
	currentLoopState := me.musicLoopState[id]

	if exists && player != nil && currentLoopState == opts.Loop {
		if opts.Exclusive {
			me.pauseAndRewindAllLocked()
		} else if player.IsPlaying() {
			return
		}

		if err := player.Rewind(); err != nil {
			log.Printf("Error rewinding player: %v", err)
		}
		player.Play()
		return
	}

	prepared, err := me.preparePlayback(audioData)
	if err != nil {
		log.Printf("Error preparing audio playback (ID: %d): %v", id, err)
		return
	}

	nextPlayer, err := me.newPlayer(prepared, opts.Loop)
	if err != nil {
		log.Printf("Error creating audio player (ID: %d): %v", id, err)
		return
	}

	if opts.Exclusive {
		me.pauseAndRewindAllLocked()
	}
	if exists && player != nil {
		player.Pause()
		if err := player.Close(); err != nil {
			log.Printf("Error closing player: %v", err)
		}
	}

	me.musicPlayers[id] = nextPlayer
	me.musicLoopState[id] = opts.Loop
	nextPlayer.Play()
}

func (me *musicEngine) preparePlayback(audioData []byte) (preparedPlayback, error) {
	if me.format == audioFormatF32 {
		return prepareF32Playback(audioData)
	}

	reader := bytes.NewReader(audioData)
	wavReader, err := wav.DecodeWithSampleRate(sampleRate, reader)
	if err != nil {
		return preparedPlayback{}, err
	}
	return newPreparedPlayback(wavReader)
}

func (me *musicEngine) newPlayer(prepared preparedPlayback, loop bool) (*audio.Player, error) {
	if me.format == audioFormatF32 {
		if loop {
			return me.audioContext.NewPlayerF32(audio.NewInfiniteLoopF32(prepared.stream, prepared.length))
		}
		return me.audioContext.NewPlayerF32(prepared.stream)
	}

	if loop {
		return me.audioContext.NewPlayer(audio.NewInfiniteLoop(prepared.stream, prepared.length))
	}
	return me.audioContext.NewPlayer(prepared.stream)
}

func (me *musicEngine) stop(id int) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if id == -1 {
		me.stopAllLocked()
		return
	}

	player, exists := me.musicPlayers[id]
	if exists && player != nil {
		player.Pause()
		if err := player.Close(); err != nil {
			log.Printf("Error closing player: %v", err)
		}
		delete(me.musicPlayers, id)
		delete(me.musicLoopState, id)
	}
}

func reloadAudioRuntimeState() {
	reloadSharedMusicData()
	for _, engine := range musicEngines.list() {
		engine.reloadAudioFiles()
	}
}
