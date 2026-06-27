package pigo8

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"sync"
)

const sampleRate = 44100

var customResourcesMu sync.RWMutex

type audioStore struct {
	mu     sync.Mutex
	loaded bool
	files  map[int][]byte
}

var sharedAudioStore = &audioStore{
	files: make(map[int][]byte),
}

func getCustomResources() *embeddedResources {
	customResourcesMu.RLock()
	defer customResourcesMu.RUnlock()
	return customResources
}

func setCustomResources(resources *embeddedResources) {
	customResourcesMu.Lock()
	customResources = resources
	customResourcesMu.Unlock()
}

func ensureSharedMusicDataLoaded() {
	sharedAudioStore.ensureLoaded()
}

func reloadSharedMusicData() {
	sharedAudioStore.reload()
}

func sharedAudioFileCount() int {
	return sharedAudioStore.fileCount()
}

func getSharedMusicData(id int) ([]byte, bool) {
	return sharedAudioStore.get(id)
}

func (s *audioStore) ensureLoaded() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadLocked()
}

func (s *audioStore) reload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files = loadRegisteredAudioData()
	s.loaded = true
}

func (s *audioStore) fileCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadLocked()
	return len(s.files)
}

func (s *audioStore) get(id int) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadLocked()
	data, ok := s.files[id]
	return data, ok
}

func (s *audioStore) loadLocked() {
	if s.loaded {
		return
	}
	s.files = loadRegisteredAudioData()
	s.loaded = true
}

func loadRegisteredAudioData() map[int][]byte {
	resources := getCustomResources()
	if resources == nil {
		log.Println("No custom resources registered, skipping audio file loading")
		return map[int][]byte{}
	}

	audioPaths, walkErr := resolveAudioPaths(resources)
	if walkErr != nil {
		log.Printf("Error walking through embedded filesystem: %v", walkErr)
	}

	musicData := make(map[int][]byte, len(audioPaths))
	for _, path := range audioPaths {
		audioID, err := parseAudioFileID(path)
		if err != nil {
			log.Printf("Warning: %v", err)
			continue
		}

		data, err := fs.ReadFile(resources.FS, path)
		if err != nil {
			log.Printf("Warning: Could not read audio file %s: %v", path, err)
			continue
		}

		if _, exists := musicData[audioID]; exists {
			log.Printf("Warning: Duplicate audio ID %d detected, overriding with %s", audioID, path)
		}

		musicData[audioID] = data
		log.Printf("Loaded audio file: %s (ID: %d)", path, audioID)
	}

	log.Printf("Loaded %d audio files", len(musicData))
	return musicData
}

func resolveAudioPaths(resources *embeddedResources) ([]string, error) {
	if len(resources.AudioPaths) > 0 {
		return uniqueSortedPaths(resources.AudioPaths), nil
	}

	var audioPaths []string
	walkErr := fs.WalkDir(resources.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, err := parseAudioFileID(path); err == nil {
			audioPaths = append(audioPaths, path)
		}
		return nil
	})
	sort.Strings(audioPaths)
	return audioPaths, walkErr
}

func uniqueSortedPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, path)
	}

	sort.Strings(unique)
	return unique
}

func parseAudioFileID(path string) (int, error) {
	filename := filepath.Base(path)
	var audioID int
	if _, err := fmt.Sscanf(filename, "music%d.wav", &audioID); err != nil {
		return 0, fmt.Errorf("audio file %q must match music%%d.wav", filename)
	}
	return audioID, nil
}
