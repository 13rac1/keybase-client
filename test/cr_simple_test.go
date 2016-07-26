// Copyright 2016 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// These tests all do one conflict-free operation while a user is unstaged.

package test

import (
	"runtime"
	"testing"
	"time"
)

// bob writes a non-conflicting file while unstaged
func TestCrUnmergedFile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			write("a/d", "uh oh"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "FILE"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "uh oh"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "FILE"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "uh oh"),
		),
	)
}

// bob writes a non-conflicting dir (containing a file) while unstaged
func TestCrUnmergedDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			write("a/d/e", "uh oh"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "DIR"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			lsdir("a/d", m{"e": "FILE"}),
			read("a/d/e", "uh oh"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "DIR"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			lsdir("a/d", m{"e": "FILE"}),
			read("a/d/e", "uh oh"),
		),
	)
}

// bob creates a non-conflicting symlink(while unstaged),
func TestCrUnmergedSymlink(t *testing.T) {
	test(t,
		skip("dokan", "Does not work with Dokan."),
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			link("a/d", "b"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "SYM"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "SYM"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
	)
}

// bob makes a non-conflicting file executable while unstaged
func TestCrUnmergedSetex(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			setex("a/b", true),
			reenableUpdates(),
			lsdir("a/", m{"b": "EXEC", "c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"b": "EXEC", "c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// bob sets the mtime on a file while unstaged
func TestCrUnmergedSetMtime(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			setmtime("a/b", targetMtime),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "world"),
			mtime("a/b", targetMtime),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "world"),
			mtime("a/b", targetMtime),
		),
	)
}

// bob sets the mtime on a moved file while unstaged
func TestCrUnmergedSetMtimeOnMovedFile(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
			mkdir("b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rename("a/b", "b/a"),
		),
		as(bob, noSync(),
			setmtime("a/b", targetMtime),
			reenableUpdates(),
			lsdir("", m{"a": "DIR", "b": "DIR"}),
			lsdir("a", m{}),
			lsdir("b", m{"a": "FILE"}),
			mtime("b/a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a": "DIR", "b": "DIR"}),
			lsdir("a", m{}),
			lsdir("b", m{"a": "FILE"}),
			mtime("b/a", targetMtime),
		),
	)
}

// bob sets the mtime on an empty file while unstaged.  We want to
// test this separately from a file with contents, to make sure we can
// properly identify a file node that is empty (and hence can be
// decoded as a DirBlock).
func TestCrUnmergedSetMtimeEmptyFile(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", ""),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "hello"),
		),
		as(bob, noSync(),
			setmtime("a/b", targetMtime),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "hello"),
			mtime("a/b", targetMtime),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "hello"),
			mtime("a/b", targetMtime),
		),
	)
}

// bob sets the mtime on a dir while unstaged
func TestCrUnmergedSetMtimeOnDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("b", "hello"),
		),
		as(bob, noSync(),
			setmtime("a", targetMtime),
			reenableUpdates(),
			lsdir("", m{"a": "DIR", "b": "FILE"}),
			read("b", "hello"),
			mtime("a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a": "DIR", "b": "FILE"}),
			read("b", "hello"),
			mtime("a", targetMtime),
		),
	)
}

// bob sets the mtime on a moved dir while unstaged
func TestCrUnmergedSetMtimeOnMovedDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
			mkdir("b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rename("a", "b/a"),
		),
		as(bob, noSync(),
			setmtime("a", targetMtime),
			reenableUpdates(),
			lsdir("", m{"b": "DIR"}),
			lsdir("b", m{"a": "DIR"}),
			mtime("b/a", targetMtime),
		),
		as(alice,
			lsdir("", m{"b": "DIR"}),
			lsdir("b", m{"a": "DIR"}),
			mtime("b/a", targetMtime),
		),
	)
}

// bob sets the mtime of a dir modified by alice.
func TestCrUnmergedSetMtimeOnModifiedDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob, noSync(),
			setmtime("a", targetMtime),
			reenableUpdates(),
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE"}),
			mtime("a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE"}),
			mtime("a", targetMtime),
		),
	)
}

// bob sets the mtime of a dir modified by both him and alice.
func TestCrUnmergedSetMtimeOnDualModifiedDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob, noSync(),
			mkfile("a/c", "hello"),
			setmtime("a", targetMtime),
			reenableUpdates(),
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE", "c$": "FILE"}),
			mtime("a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE", "c$": "FILE"}),
			mtime("a", targetMtime),
		),
	)
}

// bob deletes a non-conflicting file while unstaged
func TestCrUnmergedRmfile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			rm("a/b"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// bob deletes a non-conflicting dir while unstaged
func TestCrUnmergedRmdir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a/b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			rmdir("a/b"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// bob deletes a non-conflicting dir tree while unstaged.
// Regression for KBFS-1202.
func TestCrUnmergedRmdirTree(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a/b"),
			mkdir("a/b/c"),
			mkdir("a/b/d"),
			mkfile("a/b/c/e", "hello"),
			mkfile("a/b/d/f", "world"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/g", "merged"),
			disableUpdates(),
		),
		as(bob, noSync(),
			rm("a/b/d/f"),
			rm("a/b/c/e"),
			rmdir("a/b/d"),
			rmdir("a/b/c"),
			rmdir("a/b"),
			reenableUpdates(),
			lsdir("a/", m{"g": "FILE"}),
			read("a/g", "merged"),
		),
		as(alice, noSync(),
			mkfile("a/h", "merged2"),
			reenableUpdates(),
			lsdir("a/", m{"g": "FILE", "h": "FILE"}),
			read("a/g", "merged"),
			read("a/h", "merged2"),
		),
		as(bob,
			lsdir("a/", m{"g": "FILE", "h": "FILE"}),
			read("a/g", "merged"),
			read("a/h", "merged2"),
		),
	)
}

// bob renames a non-conflicting file while unstaged
func TestCrUnmergedRenameInDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			rename("a/b", "a/d"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE", "d": "FILE"}),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE", "d": "FILE"}),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
	)
}

// bob creates and renames a non-conflicting file while unstaged
func TestCrUnmergedCreateAndRenameInDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			write("a/b2", "hellohello"),
			rename("a/b2", "a/d"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "FILE"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hellohello"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "FILE"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hellohello"),
		),
	)
}

// bob renames a non-conflicting symlink(while unstaged),
func TestCrUnmergedRenameSymlinkInDir(t *testing.T) {
	test(t,
		skip("dokan", "Does not work with Dokan."),
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
			link("a/c", "b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/d", "world"),
		),
		as(bob, noSync(),
			rename("a/c", "a/e"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "d": "FILE", "e": "SYM"}),
			read("a/d", "world"),
			read("a/e", "hello"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "d": "FILE", "e": "SYM"}),
			read("a/d", "world"),
			read("a/e", "hello"),
		),
	)
}

// bob renames a non-conflicting file in the root dir while unstaged
func TestCrUnmergedRenameInRoot(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			rename("b", "d"),
			reenableUpdates(),
			lsdir("", m{"d": "FILE", "a": "DIR"}),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
			read("d", "hello"),
		),
		as(alice,
			lsdir("", m{"d": "FILE", "a": "DIR"}),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
			read("d", "hello"),
		),
	)
}

// bob renames a non-conflicting file across directories while unstaged
func TestCrUnmergedRenameAcrossDirs(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
			mkdir("d"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
		),
		as(bob, noSync(),
			rename("a/b", "d/e"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			lsdir("d/", m{"e": "FILE"}),
			read("a/c", "world"),
			read("d/e", "hello"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			lsdir("d/", m{"e": "FILE"}),
			read("a/c", "world"),
			read("d/e", "hello"),
		),
	)
}

// bob renames a file over an existing file
func TestCrUnmergedRenameFileOverFile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
			mkfile("a/c", "world"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/d", "just another file"),
		),
		as(bob, noSync(),
			rename("a/c", "a/b"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "d": "FILE"}),
			read("a/b", "world"),
			read("a/d", "just another file"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "d": "FILE"}),
			read("a/b", "world"),
			read("a/d", "just another file"),
		),
	)
}

// bob renames a dir over an existing empty dir
func TestCrUnmergedRenameDirOverEmptyDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a/b"),
			mkfile("a/c/d", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/e", "just another file"),
		),
		as(bob, noSync(),
			rename("a/c", "a/b"),
			reenableUpdates(),
			lsdir("a/", m{"b": "DIR", "e": "FILE"}),
			lsdir("a/b", m{"d": "FILE"}),
			read("a/b/d", "hello"),
			read("a/e", "just another file"),
		),
		as(alice,
			lsdir("a/", m{"b": "DIR", "e": "FILE"}),
			lsdir("a/b", m{"d": "FILE"}),
			read("a/b/d", "hello"),
			read("a/e", "just another file"),
		),
	)
}

// alice makes a non-conflicting dir (containing a file) while bob is
// unstaged
func TestCrMergedDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/d/e", "uh oh"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "DIR"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			lsdir("a/d", m{"e": "FILE"}),
			read("a/d/e", "uh oh"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "DIR"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			lsdir("a/d", m{"e": "FILE"}),
			read("a/d/e", "uh oh"),
		),
	)
}

// alice creates a non-conflicting symlink(while bob is unstaged),
func TestCrMergedSymlink(t *testing.T) {
	test(t,
		skip("dokan", "Does not work with Dokan."),
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			link("a/d", "b"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "SYM"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE", "d": "SYM"}),
			read("a/b", "hello"),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
	)
}

// alice makes a non-conflicting file executable while bob is unstaged
func TestCrMergedSetex(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			setex("a/b", true),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"b": "EXEC", "c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"b": "EXEC", "c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// alice set the mtime of a non-conflicting file while bob is unstaged
func TestCrMergedSetMtime(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			setmtime("a/b", targetMtime),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "world"),
			mtime("a/b", targetMtime),
		),
		as(alice,
			lsdir("a/", m{"b": "FILE", "c": "FILE"}),
			read("a/c", "world"),
			mtime("a/b", targetMtime),
		),
	)
}

// alice sets the mtime on a dir while unstaged
func TestCrMergedSetMtimeOnDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			setmtime("a", targetMtime),
		),
		as(bob, noSync(),
			write("b", "hello"),
			reenableUpdates(),
			lsdir("", m{"a": "DIR", "b": "FILE"}),
			read("b", "hello"),
			mtime("a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a": "DIR", "b": "FILE"}),
			read("b", "hello"),
			mtime("a", targetMtime),
		),
	)
}

// alice sets the mtime on a moved dir while bob is unstaged
func TestCrMergedSetMtimeOnMovedDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
			mkdir("b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			setmtime("a", targetMtime),
		),
		as(bob, noSync(),
			rename("a", "b/a"),
			reenableUpdates(),
			lsdir("", m{"b": "DIR"}),
			lsdir("b", m{"a": "DIR"}),
			mtime("b/a", targetMtime),
		),
		as(alice,
			lsdir("", m{"b": "DIR"}),
			lsdir("b", m{"a": "DIR"}),
			mtime("b/a", targetMtime),
		),
	)
}

// alice sets the mtime of a dir modified by bob.
func TestCrMergedSetMtimeOnModifiedDir(t *testing.T) {
	targetMtime := time.Now().Add(1 * time.Minute)
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			setmtime("a", targetMtime),
		),
		as(bob, noSync(),
			mkfile("a/b", "hello"),
			reenableUpdates(),
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE"}),
			mtime("a", targetMtime),
		),
		as(alice,
			lsdir("", m{"a$": "DIR"}),
			lsdir("a", m{"b$": "FILE"}),
			mtime("a", targetMtime),
		),
	)
}

// alice deletes a non-conflicting file while bob is unstaged
func TestCrMergedRmfile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rm("a/b"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// alice deletes a non-conflicting dir while bob is unstaged
func TestCrMergedRmdir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a/b"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rmdir("a/b"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
		),
	)
}

// alice deletes a non-conflicting dir tree while unstaged
func TestCrMergedRmdirTree(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a/b"),
			mkdir("a/b/c"),
			mkdir("a/b/d"),
			mkfile("a/b/c/e", "hello"),
			mkfile("a/b/d/f", "world"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rm("a/b/d/f"),
			rm("a/b/c/e"),
			rmdir("a/b/d"),
			rmdir("a/b/c"),
			rmdir("a/b"),
		),
		as(bob, noSync(),
			mkfile("a/g", "unmerged"),
			reenableUpdates(),
			lsdir("a/", m{"g": "FILE"}),
			read("a/g", "unmerged"),
		),
		as(alice,
			lsdir("a/", m{"g": "FILE"}),
			read("a/g", "unmerged"),
		),
	)
}

// alice renames a non-conflicting file while bob is unstaged
func TestCrMergedRenameInDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rename("a/b", "a/d"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE", "d": "FILE"}),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE", "d": "FILE"}),
			read("a/c", "world"),
			read("a/d", "hello"),
		),
	)
}

// alice renames a non-conflicting file in the root dir while bob is unstaged
func TestCrMergedRenameInRoot(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rename("b", "d"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("", m{"d": "FILE", "a": "DIR"}),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
			read("d", "hello"),
		),
		as(alice,
			lsdir("", m{"d": "FILE", "a": "DIR"}),
			lsdir("a/", m{"c": "FILE"}),
			read("a/c", "world"),
			read("d", "hello"),
		),
	)
}

// alice renames a non-conflicting file across directories while bob
// is unstaged
func TestCrMergedRenameAcrossDirs(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
			mkdir("d"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			rename("a/b", "d/e"),
		),
		as(bob, noSync(),
			write("a/c", "world"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
			lsdir("d/", m{"e": "FILE"}),
			read("a/c", "world"),
			read("d/e", "hello"),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
			lsdir("d/", m{"e": "FILE"}),
			read("a/c", "world"),
			read("d/e", "hello"),
		),
	)
}

// alice and bob write(the same dir (containing a file) while bob's unstaged),
func TestCrMergeDir(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/b/c", "hello"),
		),
		as(bob, noSync(),
			write("a/b/d", "world"),
			reenableUpdates(),
			lsdir("a/", m{"b": "DIR"}),
			lsdir("a/b", m{"c": "FILE", "d": "FILE"}),
			read("a/b/c", "hello"),
			read("a/b/d", "world"),
		),
		as(alice,
			lsdir("a/", m{"b": "DIR"}),
			lsdir("a/b", m{"c": "FILE", "d": "FILE"}),
			read("a/b/c", "hello"),
			read("a/b/d", "world"),
		),
	)
}

// alice and bob both delete the same file
func TestCrUnmergedBothRmfile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/c", "world"),
			rm("a/b"),
		),
		as(bob, noSync(),
			rm("a/b"),
			reenableUpdates(),
			lsdir("a/", m{"c": "FILE"}),
		),
		as(alice,
			lsdir("a/", m{"c": "FILE"}),
		),
	)
}

// bob exclusively creates a file while on an unmerged branch.
func TestCrCreateFileExclOnStaged(t *testing.T) {
	var skipOp optionOp
	if runtime.GOOS == "darwin" {
		skipOp = skip("fuse", "osxfuse doesn't pass through O_EXCL yet")
	} else {
		skipOp = func(c *opt) {}
	}
	test(t,
		skipOp,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob, noSync(),
			mkfileexcl("a/c"),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE", "c$": "FILE"}),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE", "c$": "FILE"}),
		),
	)
}

// alice and bob both exclusively create the same file, but neither write to
// it. Since the creates are exclusive, only the winning one (alice) should
// succeed.
func TestCrBothCreateFileExcl(t *testing.T) {
	var skipOp optionOp
	if runtime.GOOS == "darwin" {
		skipOp = skip("fuse", "osxfuse doesn't pass through O_EXCL yet")
	} else {
		skipOp = func(c *opt) {}
	}
	test(t,
		skipOp,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfileexcl("a/b"),
		),
		as(bob, noSync(),
			expectError(mkfileexcl("a/b"), "b already exists"),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
		),
	)
}

// alice and bob both exclusively create the same file, but neither write to
// it. This test is run in parallel. Bob's exclusive create is stalled on MD's
// Put. After stall happens, alice creates the file. This makes sure Alice's
// exclusive create happens precisely before Bob's MD Put.
func TestCrBothCreateFileExclParallel(t *testing.T) {
	var skipOp optionOp
	if runtime.GOOS == "darwin" {
		skipOp = skip("fuse", "osxfuse doesn't pass through O_EXCL yet")
	} else {
		skipOp = func(c *opt) {}
	}
	test(t,
		skipOp,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			lsdir("a/", m{}),
		),
		as(bob, stallOnMDPut()),
		parallel(
			as(bob,
				expectError(mkfileexcl("a/b"), "b already exists"),
				lsdir("a/", m{"b$": "FILE"}),
			),
			sequential(
				as(bob, noSync(), waitForStalledMDPut()),
				as(alice,
					mkfileexcl("a/b"),
					lsdir("a/", m{"b$": "FILE"}),
				),
				as(bob, noSync(), undoStallOnMDPut()),
			),
		),
	)
}

// alice and bob both create the same file, but neither write to it
func TestCrBothCreateFile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", ""),
		),
		as(bob, noSync(),
			mkfile("a/b", ""),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
	)
}

// alice and bob both create the same file, and alice wrote to it
func TestCrBothCreateFileMergedWrite(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob, noSync(),
			mkfile("a/b", ""),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hello"),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hello"),
		),
	)
}

// alice and bob both create the same file, and alice truncated to it
func TestCrBothCreateFileMergedTruncate(t *testing.T) {
	const flen = 401001
	fdata := string(make([]byte, flen))
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", ""),
			truncate("a/b", flen),
		),
		as(bob, noSync(),
			mkfile("a/b", ""),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", fdata),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", fdata),
		),
	)
}

// alice and bob both create the same file, and bob wrote to it
func TestCrBothCreateFileUnmergedWrite(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", ""),
		),
		as(bob, noSync(),
			mkfile("a/b", "hello"),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hello"),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hello"),
		),
	)
}

// alice and bob both create the same file, and bob truncated to it
func TestCrBothCreateFileUnmergedTruncate(t *testing.T) {
	const flen = 401001
	fdata := string(make([]byte, flen))
	test(t,
		users("alice", "bob"),
		as(alice,
			mkdir("a"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			mkfile("a/b", ""),
		),
		as(bob, noSync(),
			mkfile("a/b", ""),
			truncate("a/b", flen),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", fdata),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", fdata),
		),
	)
}

// alice and bob both truncate the same file
func TestCrBothTruncateFile(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			truncate("a/b", 0),
		),
		as(bob, noSync(),
			truncate("a/b", 0),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
	)
}

// alice and bob both truncate the same file to a non-zero size
func TestCrBothTruncateFileNonZero(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			truncate("a/b", 4),
		),
		as(bob, noSync(),
			truncate("a/b", 4),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hell"),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", "hell"),
		),
	)
}

// alice and bob both truncate the same file, and alice wrote to it first
func TestCrBothTruncateFileMergedWrite(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			write("a/b", "world"),
			truncate("a/b", 0),
		),
		as(bob, noSync(),
			truncate("a/b", 0),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
	)
}

// alice and bob both truncate the same file, and bob wrote to first
func TestCrBothTruncateFileUnmergedWrite(t *testing.T) {
	test(t,
		users("alice", "bob"),
		as(alice,
			mkfile("a/b", "hello"),
		),
		as(bob,
			disableUpdates(),
		),
		as(alice,
			truncate("a/b", 0),
		),
		as(bob, noSync(),
			write("a/b", "world"),
			truncate("a/b", 0),
			reenableUpdates(),
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
		as(alice,
			lsdir("a/", m{"b$": "FILE"}),
			read("a/b", ""),
		),
	)
}
