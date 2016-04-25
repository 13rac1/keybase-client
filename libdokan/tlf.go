// Copyright 2016 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// +build windows

package libdokan

import (
	"sync"

	"github.com/keybase/kbfs/dokan"
	"github.com/keybase/kbfs/libkbfs"
	"golang.org/x/net/context"
)

// TLF represents the root directory of a TLF. It wraps a lazy-loaded
// Dir.
type TLF struct {
	refcount refcount

	folder *Folder

	dirLock sync.RWMutex
	dir     *Dir

	emptyFile
}

func newTLF(fl *FolderList, h *libkbfs.TlfHandle) *TLF {
	folder := newFolder(fl, h)
	tlf := &TLF{
		folder: folder,
	}
	tlf.refcount.Increase()
	return tlf
}

func (tlf *TLF) isPublic() bool {
	return tlf.folder.list.public
}

func (tlf *TLF) getStoredDir() *Dir {
	tlf.dirLock.RLock()
	defer tlf.dirLock.RUnlock()
	return tlf.dir
}

func (tlf *TLF) filterEarlyExitError(ctx context.Context, err error) (
	exitEarly bool, retErr error) {
	switch err := err.(type) {
	case nil:
		// No error.
		return false, nil

	case libkbfs.WriteAccessError:
		// No permission to create TLF, so pretend it's still
		// empty.
		//
		// In theory, we need to invalidate this once the TLF
		// is created, but in practice, the Linux kernel
		// doesn't cache readdir results, and probably not
		// OSXFUSE either.
		tlf.folder.fs.log.CDebugf(ctx,
			"No permission to write to %s, so pretending it's empty",
			tlf.folder.name)
		return true, nil

	case libkbfs.MDServerErrorWriteAccess:
		// Same as above; cannot fallthrough in type switch
		tlf.folder.fs.log.CDebugf(ctx,
			"No permission to write to %s, so pretending it's empty",
			tlf.folder.name)
		return true, nil

	default:
		// Some other error.
		return true, err
	}
}

func (tlf *TLF) loadDirHelper(ctx context.Context, info string, filterErr bool) (dir *Dir, exitEarly bool, err error) {
	dir = tlf.getStoredDir()
	if dir != nil {
		return dir, false, nil
	}

	tlf.dirLock.Lock()
	defer tlf.dirLock.Unlock()
	// Need to check for nilness again to avoid racing with other
	// calls to loadDir().
	if tlf.dir != nil {
		return tlf.dir, false, nil
	}

	name := tlf.folder.name()

	tlf.folder.fs.log.CDebugf(ctx, "Loading root directory for folder %s "+
		"(public: %t) for %s", name, tlf.isPublic(), info)
	defer func() {
		if filterErr {
			exitEarly, err = tlf.filterEarlyExitError(ctx, err)
		}

		tlf.folder.reportErr(ctx, libkbfs.ReadMode, err, nil)
	}()

	rootNode, _, err :=
		tlf.folder.fs.config.KBFSOps().GetOrCreateRootNode(
			ctx, tlf.folder.h, libkbfs.MasterBranch)
	if err != nil {
		return nil, false, err
	}

	err = tlf.folder.setFolderBranch(rootNode.GetFolderBranch())
	if err != nil {
		return nil, false, err
	}

	tlf.folder.nodes[rootNode.GetID()] = tlf
	tlf.dir = newDir(tlf.folder, rootNode, string(name), nil)
	// TLFs should be cached.
	tlf.dir.refcount.Increase()
	tlf.folder.lockedAddNode(rootNode, tlf.dir)

	return tlf.dir, false, nil
}

// loadDirAllowNonexistent loads a TLF if it's not already loaded.  If
// the TLF doesn't yet exist, it still returns a nil error and
// indicates that the calling function should pretend it's an empty
// folder.
func (tlf *TLF) loadDirAllowNonexistent(ctx context.Context, info string) (
	*Dir, bool, error) {
	return tlf.loadDirHelper(ctx, info, true)
}

// GetFileInformation for dokan.
func (tlf *TLF) GetFileInformation(fi *dokan.FileInfo) (st *dokan.Stat, err error) {
	dir := tlf.getStoredDir()
	if dir == nil {
		return defaultDirectoryInformation()
	}

	return dir.GetFileInformation(fi)
}

// open tries to open a file.
func (tlf *TLF) open(ctx context.Context, oc *openContext, path []string) (dokan.File, bool, error) {
	if len(path) == 0 {
		if oc.mayNotBeDirectory() {
			return nil, true, dokan.ErrFileIsADirectory
		}
		tlf.refcount.Increase()
		return tlf, true, nil
	}
	dir, exitEarly, err := tlf.loadDirAllowNonexistent(ctx, "open")
	if err != nil {
		return nil, false, err
	}
	if exitEarly {
		return nil, false, dokan.ErrObjectNameNotFound
	}
	return dir.open(ctx, oc, path)
}

// FindFiles does readdir for dokan.
func (tlf *TLF) FindFiles(fi *dokan.FileInfo, callback func(*dokan.NamedStat) error) (err error) {
	ctx, cancel := NewContextWithOpID(tlf.folder.fs, "TLF FindFiles")
	defer cancel()
	dir, exitEarly, err := tlf.loadDirAllowNonexistent(ctx, "FindFiles")
	if err != nil {
		return errToDokan(err)
	}
	if exitEarly {
		return dokan.ErrObjectNameNotFound
	}
	return dir.FindFiles(fi, callback)
}

// CanDeleteDirectory - return just nil because tlfs
// can always be removed from favorites.
func (tlf *TLF) CanDeleteDirectory(*dokan.FileInfo) (err error) {
	return nil
}

// Cleanup - forget references, perform deletions etc.
func (tlf *TLF) Cleanup(fi *dokan.FileInfo) {
	var err error
	if fi != nil && fi.DeleteOnClose() {
		ctx, cancel := NewContextWithOpID(tlf.folder.fs, "TLF Cleanup")
		defer tlf.folder.reportErr(ctx, libkbfs.WriteMode, err, cancel)
		err = tlf.folder.fs.config.KBFSOps().DeleteFavorite(ctx,
			string(tlf.folder.name()), tlf.isPublic())
	}

	if tlf.refcount.Decrease() {
		dir := tlf.getStoredDir()
		if dir == nil {
			return
		}
		dir.Cleanup(fi)
	}
}
