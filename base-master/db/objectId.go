package db

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type ObjectId string
func (obj ObjectId) String() string {
	return string(obj)
}
type Base struct {
	Id ObjectId `bson:"_id,omitempty" json:",omitempty"`
}

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineId = readMachineId()
var processId = os.Getpid()

// objectIdCounter is atomically incremented when generating a new ObjectId
// using NewObjectId() function. It's used as a counter part of an id.
var objectIdCounter = readRandomUint32()

// NewObjectId returns a new unique ObjectId.
// rccode 资源类型
func NewObjectId(rc uint16) ObjectId {
	var b [16]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(processId >> 8)
	b[8] = byte(processId)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	// platform, 2 bytes
	// TODO
	// type, 2 bytes
	binary.BigEndian.PutUint16(b[14:], rc)

	return ObjectId(hex.EncodeToString(b[:]))
}

func NewStringObjectId(rc uint16) string {
	return string(NewObjectId(rc))
}

// readMachineId generates and returns a machine id.
// If this function fails to get the hostname it will cause a runtime error.
func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// readRandomUint32 returns a random objectIdCounter.
func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot read random object id: %v", err))
	}
	return uint32((uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24))
}

// Hex returns a hex representation of the ObjectId.
//func (id ObjectId) Hex() string {
//	return hex.EncodeToString([]byte(id))
//}

// ObjectIdHex returns an ObjectId from the provided hex representation.
// Calling this function with an invalid hex representation will
// cause a runtime panic. See the IsObjectIdHex function.
//func ObjectIdHex(s string) ObjectId {
//	d, err := hex.DecodeString(s)
//	if err != nil || len(d) != 16 {
//		panic(fmt.Sprintf("invalid input to ObjectIdHex: %q", s))
//	}
//	return ObjectId(d)
//}

// 解析no
func (id ObjectId) ParseNo() (int, error) {
	if len(id) != 32 {
		return 0, fmt.Errorf("invalid ObjectId: %s", id)
	}

	no, err := strconv.ParseInt(string(id[18:24]), 16, 0)
	return int(no), err
}
