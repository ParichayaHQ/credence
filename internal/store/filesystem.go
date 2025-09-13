package store

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FilesystemBlobStore implements BlobStore using filesystem storage
type FilesystemBlobStore struct {
	config *BlobStoreConfig
	basePath string
	
	// Metadata for statistics
	meta     *FSMetadata
	metaFile string
	metaMu   sync.RWMutex
	
	// Cleanup
	cleanupTicker *time.Ticker
	cleanupStop   chan struct{}
	
	// State
	mu     sync.RWMutex
	closed bool
}

// FSMetadata tracks filesystem store metadata
type FSMetadata struct {
	TotalBlobs   int64     `json:"total_blobs"`
	TotalBytes   int64     `json:"total_bytes"`
	LastAccessed time.Time `json:"last_accessed"`
	LastCleanup  time.Time `json:"last_cleanup"`
	
	// Index of CID -> file info for quick stats
	Files map[string]FSFileInfo `json:"files"`
}

// FSFileInfo tracks individual file metadata
type FSFileInfo struct {
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`
}

// NewFilesystemBlobStore creates a new filesystem-based blob store
func NewFilesystemBlobStore(config *BlobStoreConfig) (*FilesystemBlobStore, error) {
	if config.FSPath == "" {
		return nil, ErrInvalidConfig("fs_path required for filesystem backend")
	}
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.FSPath, 0755); err != nil {
		return nil, ErrDatabase("mkdir", err)
	}
	
	store := &FilesystemBlobStore{
		config:   config,
		basePath: config.FSPath,
		metaFile: filepath.Join(config.FSPath, ".metadata.json"),
		cleanupStop: make(chan struct{}),
	}
	
	// Load or create metadata
	if err := store.loadMetadata(); err != nil {
		return nil, err
	}
	
	// Start cleanup routine if enabled
	if config.EnableCleanup {
		store.cleanupTicker = time.NewTicker(config.CleanupInterval)
		go store.cleanupRoutine()
	}
	
	return store, nil
}

// loadMetadata loads or initializes metadata
func (f *FilesystemBlobStore) loadMetadata() error {
	f.meta = &FSMetadata{
		Files: make(map[string]FSFileInfo),
	}
	
	// Try to load existing metadata
	if data, err := os.ReadFile(f.metaFile); err == nil {
		if err := json.Unmarshal(data, f.meta); err != nil {
			// If metadata is corrupted, rebuild from filesystem
			return f.rebuildMetadata()
		}
	} else if !os.IsNotExist(err) {
		return ErrDatabase("load_metadata", err)
	}
	
	return nil
}

// saveMetadata saves metadata to disk
func (f *FilesystemBlobStore) saveMetadata() error {
	data, err := json.MarshalIndent(f.meta, "", "  ")
	if err != nil {
		return ErrDatabase("marshal_metadata", err)
	}
	
	return os.WriteFile(f.metaFile, data, 0644)
}

// rebuildMetadata rebuilds metadata from filesystem scan
func (f *FilesystemBlobStore) rebuildMetadata() error {
	f.meta = &FSMetadata{
		Files: make(map[string]FSFileInfo),
	}
	
	err := filepath.WalkDir(f.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		
		// Extract CID from filename
		relPath, _ := filepath.Rel(f.basePath, path)
		cid := strings.ReplaceAll(relPath, string(filepath.Separator), "")
		
		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		f.meta.Files[cid] = FSFileInfo{
			Size:       info.Size(),
			CreatedAt:  info.ModTime(),
			AccessedAt: info.ModTime(),
		}
		
		f.meta.TotalBlobs++
		f.meta.TotalBytes += info.Size()
		
		return nil
	})
	
	if err != nil {
		return ErrDatabase("rebuild_metadata", err)
	}
	
	return f.saveMetadata()
}

// cidToPath converts a CID to a filesystem path with directory structure
func (f *FilesystemBlobStore) cidToPath(cid string) string {
	// Create 2-level directory structure from CID prefix
	// Example: QmAbc123... -> Q/m/QmAbc123...
	if len(cid) < 4 {
		return filepath.Join(f.basePath, "short", cid)
	}
	
	return filepath.Join(f.basePath, cid[:1], cid[1:2], cid)
}

// Store implements BlobStore.Store
func (f *FilesystemBlobStore) Store(ctx context.Context, data []byte) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if f.closed {
		return "", ErrClosed
	}
	
	if int64(len(data)) > f.config.MaxBlobSize {
		return "", ErrTooLarge
	}
	
	// Generate CID using SHA-256 hash
	cid, err := generateCID(data)
	if err != nil {
		return "", ErrDatabase("cid_generation", err)
	}
	
	// Check if file already exists
	filePath := f.cidToPath(cid)
	if _, err := os.Stat(filePath); err == nil {
		// File exists, update access time in metadata
		f.updateAccessTime(cid)
		return cid, nil
	}
	
	// Create directory structure
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", ErrDatabase("mkdir", err)
	}
	
	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", ErrDatabase("write_file", err)
	}
	
	// Update metadata
	f.updateMetadata(cid, int64(len(data)), true)
	
	return cid, nil
}

// Get implements BlobStore.Get
func (f *FilesystemBlobStore) Get(ctx context.Context, cid string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if f.closed {
		return nil, ErrClosed
	}
	
	filePath := f.cidToPath(cid)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFoundCID(cid)
		}
		return nil, ErrDatabase("read_file", err)
	}
	
	// Update access time
	f.updateAccessTime(cid)
	
	return data, nil
}

// Has implements BlobStore.Has
func (f *FilesystemBlobStore) Has(ctx context.Context, cid string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if f.closed {
		return false, ErrClosed
	}
	
	filePath := f.cidToPath(cid)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, ErrDatabase("stat_file", err)
	}
	
	return true, nil
}

// Delete implements BlobStore.Delete
func (f *FilesystemBlobStore) Delete(ctx context.Context, cid string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if f.closed {
		return ErrClosed
	}
	
	filePath := f.cidToPath(cid)
	
	// Get file size for metadata update
	var size int64
	if info, err := os.Stat(filePath); err == nil {
		size = info.Size()
	}
	
	// Delete file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return ErrDatabase("delete_file", err)
	}
	
	// Update metadata
	f.removeFromMetadata(cid, size)
	
	// Try to remove empty directories
	f.cleanupEmptyDirs(filepath.Dir(filePath))
	
	return nil
}

// Stats implements BlobStore.Stats
func (f *FilesystemBlobStore) Stats() BlobStats {
	f.metaMu.RLock()
	defer f.metaMu.RUnlock()
	
	return BlobStats{
		TotalBlobs:   f.meta.TotalBlobs,
		TotalBytes:   f.meta.TotalBytes,
		LastAccessed: f.meta.LastAccessed,
	}
}

// updateMetadata updates metadata for a stored blob
func (f *FilesystemBlobStore) updateMetadata(cid string, size int64, isNew bool) {
	f.metaMu.Lock()
	defer f.metaMu.Unlock()
	
	now := time.Now()
	
	if isNew {
		f.meta.Files[cid] = FSFileInfo{
			Size:       size,
			CreatedAt:  now,
			AccessedAt: now,
		}
		f.meta.TotalBlobs++
		f.meta.TotalBytes += size
	} else {
		if info, exists := f.meta.Files[cid]; exists {
			info.AccessedAt = now
			f.meta.Files[cid] = info
		}
	}
	
	f.meta.LastAccessed = now
	
	// Don't save metadata in background during rapid operations
	// This prevents race conditions during testing
}

// updateAccessTime updates the access time for a blob
func (f *FilesystemBlobStore) updateAccessTime(cid string) {
	f.updateMetadata(cid, 0, false)
}

// removeFromMetadata removes a blob from metadata
func (f *FilesystemBlobStore) removeFromMetadata(cid string, size int64) {
	f.metaMu.Lock()
	defer f.metaMu.Unlock()
	
	if _, exists := f.meta.Files[cid]; exists {
		delete(f.meta.Files, cid)
		f.meta.TotalBlobs--
		f.meta.TotalBytes -= size
		
		// Metadata will be saved on close
	}
}

// cleanupEmptyDirs removes empty directories up the tree
func (f *FilesystemBlobStore) cleanupEmptyDirs(dir string) {
	if dir == f.basePath {
		return
	}
	
	// Check if directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	
	// Remove empty directory
	os.Remove(dir)
	
	// Recursively check parent
	f.cleanupEmptyDirs(filepath.Dir(dir))
}

// cleanupRoutine periodically cleans up old files
func (f *FilesystemBlobStore) cleanupRoutine() {
	for {
		select {
		case <-f.cleanupTicker.C:
			f.runCleanup()
		case <-f.cleanupStop:
			return
		}
	}
}

// runCleanup removes files older than MaxAge
func (f *FilesystemBlobStore) runCleanup() {
	if f.config.MaxAge <= 0 {
		return
	}
	
	f.metaMu.Lock()
	defer f.metaMu.Unlock()
	
	cutoff := time.Now().Add(-f.config.MaxAge)
	var toDelete []string
	
	// Find old files
	for cid, info := range f.meta.Files {
		if info.AccessedAt.Before(cutoff) {
			toDelete = append(toDelete, cid)
			if len(toDelete) >= f.config.CleanupBatchSize {
				break
			}
		}
	}
	
	// Delete old files
	for _, cid := range toDelete {
		filePath := f.cidToPath(cid)
		if err := os.Remove(filePath); err == nil {
			if info, exists := f.meta.Files[cid]; exists {
				delete(f.meta.Files, cid)
				f.meta.TotalBlobs--
				f.meta.TotalBytes -= info.Size
			}
		}
	}
	
	if len(toDelete) > 0 {
		f.meta.LastCleanup = time.Now()
		// Metadata will be saved on close
	}
}

// Close implements BlobStore.Close
func (f *FilesystemBlobStore) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if f.closed {
		return nil
	}
	
	// Stop cleanup routine
	if f.cleanupTicker != nil {
		f.cleanupTicker.Stop()
		close(f.cleanupStop)
	}
	
	// Save final metadata
	f.saveMetadata()
	
	f.closed = true
	return nil
}