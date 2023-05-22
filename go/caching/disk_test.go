package caching

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInitDiskCache(t *testing.T) {
	tests := []struct {
		name string
		want *LocalDiskCache
	}{
		{
			name: "successful init",
			want: &LocalDiskCache{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitDiskCache(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitDiskCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalDiskCache_Read(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	defer os.Remove(tempFile.Name())

	// Write some content to the temporary file
	content := []byte("test content")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write content to temporary file: %v", err)
	}

	tempFile.Close()

	cache := LocalDiskCache{}

	t.Run("Read existing file successfully", func(t *testing.T) {
		file, err := cache.Read(tempFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if string(file) != string(content) {
			t.Errorf("Expected file content to match, got: %s", string(file))
		}
	})

	t.Run("Read non-existing file", func(t *testing.T) {
		_, err := cache.Read("non_existing_file")
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected os.ErrNotExist error, got: %v", err)
		}
	})

	t.Run("Read empty file successfully", func(t *testing.T) {
		emptyFile, err := os.CreateTemp("", "emptyfile")
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		defer os.Remove(emptyFile.Name())

		emptyFile.Close()

		file, err := cache.Read(emptyFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(file) != 0 {
			t.Errorf("Expected empty file content, got: %s", string(file))
		}
	})

	t.Run("Read large file successfully", func(t *testing.T) {
		largeContent := make([]byte, 1024*1024*10) // 10 MB
		largeFile, err := os.CreateTemp("", "largefile")
		if err != nil {
			t.Fatalf("Failed to create large file: %v", err)
		}

		defer os.Remove(largeFile.Name())

		if _, err := largeFile.Write(largeContent); err != nil {
			t.Fatalf("Failed to write content to large file: %v", err)
		}

		largeFile.Close()

		file, err := cache.Read(largeFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if string(file) != string(largeContent) {
			t.Errorf("Expected file content to match, got: %s", string(file))
		}
	})
}

func TestLocalDiskCache_Write(t *testing.T) {
	cache := LocalDiskCache{}

	t.Run("Write file in non-existing directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}

		defer os.RemoveAll(tempDir)

		filepath := filepath.Join(tempDir, "newfile.txt")
		data := []byte("test data")

		err = cache.Write(filepath, data)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify that the file was created
		if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected file to exist, but it doesn't")
		}
	})

	t.Run("Write file in existing directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		filepath := filepath.Join(tempDir, "newfile.txt")
		data := []byte("test data")

		// Create the directory manually
		err = os.MkdirAll(tempDir, 0700)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err = cache.Write(filepath, data)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify that the file was created
		if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected file to exist, but it doesn't")
		}
	})

	t.Run("Write file with write error", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		filepath := tempFile.Name()
		data := []byte("test data")

		// Set the file permissions to read-only to trigger a write error
		err = os.Chmod(filepath, 0400)
		if err != nil {
			t.Fatalf("Failed to set file permissions: %v", err)
		}

		err = cache.Write(filepath, data)
		if err == nil {
			t.Error("Expected a write error, but no error occurred")
		}
	})
}
