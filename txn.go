package mdb

/*
#cgo LDFLAGS: -L/usr/local/lib -llmdb
#cgo CFLAGS: -I/usr/local

#include <stdlib.h>
#include <stdio.h>
#include <lmdb.h>
*/
import "C"

import (
	"unsafe"
)

// DBIOpen Database Flags
const (
	REVERSEKEY  = C.MDB_REVERSEKEY // use reverse string keys
	DUPSORT     = C.MDB_DUPSORT // use sorted duplicates
	INTEGERKEY = C.MDB_INTEGERKEY // numeric keys in native byte order. The keys must all be of the same size.
	DUPFIXED = C.MDB_DUPFIXED // with DUPSORT, sorted dup items have fixed size
	INTEGERDUP = C.MDB_INTEGERDUP // with DUPSORT, dups are numeric in native byte order
	REVERSEDUP = C.MDB_REVERSEDUP // with DUPSORT, use reverse string dups 
	CREATE = C.MDB_CREATE // create DB if not already existing
)

// put flags
const (
	NODUPDATA = C.MDB_NODUPDATA
	DUPSORT = C.MDB_DUBSORT
	NOOVERWRITE = C.MDB_NOOVERWRITE
	RESERVE = C.MDB_RESERVE
	APPEND = C.MDB_APPEND
	APPENDDUP = C.MDB_APPENDDUP
)

// Txn is Opaque structure for a transaction handle. 
// All database operations require a transaction handle. 
// Transactions may be read-only or read-write.
type Txn struct {
	_txn *C.MDB_txn
}

func (env *Env) BeginTxn(parent *Txn, flags uint) (* Txn, error) {
	var _txn *C.MDB_txn
	ret := C.mdb_txn_begin(env._env, parent._txn, C.uint(flags), &_txn)
	if ret != SUCCESS {
		return nil, Errno(ret)
	}
	return &Txn{_txn}, nil
}

func (txn *Txn) Commit() error {
	ret := C.mdb_txn_commit(txn._txn)
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}

func (txn *Txn) Abort() {
	C.mdb_txn_abort(txn._txn)
	txn._txn = nil
}

func (txn *Txn) Reset() {
	C.mdb_txn_reset(txn._txn)
}

func (txn *Txn) Renew() error {
	ret := C.mdb_txn_renew(txn._txn)
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}

func (txn *Txn) DBIOpen(name string, flags uint) (DBI, error) {
	var _dbi C.MDB_dbi
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.mdb_dbi_open(txn._txn, cname, C.uint(flags), &_dbi)
	if ret != SUCCESS {
		return 0, Errno(ret)
	}
	return DBI(_dbi), nil
}

func (txn *Txn) Stat(dbi DBI) (*Stat, error) {
	var _stat C.MDB_stat
	ret := C.mdb_stat(txn._txn, C.MDB_dbi(dbi), &_stat)
	if ret != SUCCESS {
		return nil, Errno(ret)
	}
	stat := Stat{PSize: uint(_stat.ms_psize),
		Depth:         uint(_stat.ms_depth),
		BranchPages:   uint64(_stat.ms_branch_pages),
		LeafPages:     uint64(_stat.ms_leaf_pages),
		OwerflowPages: uint64(_stat.ms_overflow_pages),
		Entries:       uint64(_stat.ms_entries)}
	return &stat, nil
}

func (txn *Txn) Drop(dbi DBI, del int) error {
	ret := C.mdb_drop(txn._txn, C.MDB_dbi(dbi), C.int(del))
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}

func (txn *Txn) Get(dbi DBI, key []byte) ([]byte, error) {
	var ckey *C.MDB_val
	ckey.mv_size = C.size_t(len(key))
	ckey.mv_data = unsafe.Pointer(&key[0])
	var cval *C.MDB_val
	ret := C.mdb_get(txn._txn, C.MDB_dbi(dbi), ckey, cval)
	if ret != SUCCESS {
		return nil, Errno(ret)
	}
	val = C.GoBytes(cval.mv_data, cval.mv_size)
	return val, nil
}

func (txn *Txn) Put(dbi DBI, key []byte, val []byte, flags uint) error {
	var ckey *C.MDB_val
	ckey.mv_size = C.size_t(len(key))
	ckey.mv_data = unsafe.Pointer(&key[0])
	var cval *C.MDB_val
	cval.mv_size = C.size_t(len(val))
	cval.mv_data = unsafe.Pointer(&val[0])
	ret := C.mdb_put(txn._txn, C.MDB_dbi(dbi), ckey, cval, C.uint(flags))
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}

func (txn *Txn) Del(dbi DBI, key []byte, val []byte) error {
	var ckey *C.MDB_val
	ckey.mv_size = C.size_t(len(key))
	ckey.mv_data = unsafe.Pointer(&key[0])
	// XXX null val?
	var cval *C.MDB_val
	cval.mv_size = C.size_t(len(val))
	cval.mv_data = unsafe.Pointer(&val[0])
	ret := C.mdb_del(txn._txn, C.MDB_dbi(dbi), ckey, cval)
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}

type Cursor struct {
	_cursor *C.MDB_cursor
}

func (txn *) CursorOpen(dbi DBI) (*Cursor, error) {
	var _cursor *C.MDB_cursor
	ret := C.mdb_cursor_open(txn._txn, C.MDB_dbi(dbi), &_cursor)
	if ret != SUCCESS {
		return nil, Errno(ret)
	}
	return &Cursor{_cursor}, nil
}

func (txn *) CursorRenew(cursor *Cursor) error {
	ret := C.mdb_cursor_renew(txn._txn, cursor._cursor)
	if ret != SUCCESS {
		return Errno(ret)
	}
	return nil
}
