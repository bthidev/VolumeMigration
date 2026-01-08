package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupLocal(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files and subdirectories
	testFile := filepath.Join(tempDir, "test.txt")
	testSubDir := filepath.Join(tempDir, "subdir")

	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.MkdirAll(testSubDir, 0755); err != nil {
		t.Fatalf("Failed to create test subdirectory: %v", err)
	}

	testSubFile := filepath.Join(testSubDir, "subfile.txt")
	if err := os.WriteFile(testSubFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create test subfile: %v", err)
	}

	// Verify files exist before cleanup
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Test file should exist before cleanup")
	}
	if _, err := os.Stat(testSubFile); os.IsNotExist(err) {
		t.Fatal("Test subfile should exist before cleanup")
	}

	// Perform cleanup
	if err := CleanupLocal(tempDir); err != nil {
		t.Errorf("CleanupLocal() failed: %v", err)
	}

	// Verify directory and all contents are removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp directory should not exist after cleanup")
	}
}

func TestCleanupLocal_NonExistentDirectory(t *testing.T) {
	// Try to cleanup a directory that doesn't exist
	nonExistentDir := "/tmp/nonexistent-test-dir-12345"

	// Should not return error for non-existent directory (os.RemoveAll handles this gracefully)
	if err := CleanupLocal(nonExistentDir); err != nil {
		t.Errorf("CleanupLocal() should not error on non-existent directory: %v", err)
	}
}

func TestCleanupArchives(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create test archive files
	archivePaths := map[string]string{
		"volume1": filepath.Join(tempDir, "volume1.tar.gz"),
		"volume2": filepath.Join(tempDir, "volume2.tar.gz"),
		"volume3": filepath.Join(tempDir, "volume3.tar.gz"),
	}

	// Create the archive files
	for _, path := range archivePaths {
		if err := os.WriteFile(path, []byte("archive content"), 0644); err != nil {
			t.Fatalf("Failed to create test archive: %v", err)
		}
	}

	// Verify files exist
	for volumeName, path := range archivePaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Archive for %s should exist before cleanup", volumeName)
		}
	}

	// Perform cleanup
	if err := CleanupArchives(archivePaths); err != nil {
		t.Errorf("CleanupArchives() failed: %v", err)
	}

	// Verify all archives are removed
	for volumeName, path := range archivePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Archive for %s should not exist after cleanup", volumeName)
		}
	}
}

func TestCleanupArchives_PartialCleanup(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create test archive files
	archivePaths := map[string]string{
		"volume1": filepath.Join(tempDir, "volume1.tar.gz"),
		"volume2": filepath.Join(tempDir, "volume2.tar.gz"),
	}

	// Create only the first archive
	if err := os.WriteFile(archivePaths["volume1"], []byte("archive content"), 0644); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Don't create volume2 - it doesn't exist

	// Perform cleanup - should handle missing files gracefully
	if err := CleanupArchives(archivePaths); err != nil {
		t.Errorf("CleanupArchives() should handle missing files gracefully: %v", err)
	}

	// Verify volume1 is removed
	if _, err := os.Stat(archivePaths["volume1"]); !os.IsNotExist(err) {
		t.Error("Archive for volume1 should not exist after cleanup")
	}
}

func TestCleanupArchives_EmptyMap(t *testing.T) {
	// Test with empty archive map
	archivePaths := map[string]string{}

	if err := CleanupArchives(archivePaths); err != nil {
		t.Errorf("CleanupArchives() should handle empty map: %v", err)
	}
}

func TestCleanupLocal_NestedDirectories(t *testing.T) {
	// Create a complex nested directory structure
	tempDir := t.TempDir()

	nestedPath := filepath.Join(tempDir, "level1", "level2", "level3")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create files at various levels
	files := []string{
		filepath.Join(tempDir, "root.txt"),
		filepath.Join(tempDir, "level1", "file1.txt"),
		filepath.Join(tempDir, "level1", "level2", "file2.txt"),
		filepath.Join(nestedPath, "file3.txt"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Cleanup
	if err := CleanupLocal(tempDir); err != nil {
		t.Errorf("CleanupLocal() failed on nested structure: %v", err)
	}

	// Verify everything is removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Nested directory structure should be completely removed")
	}
}
