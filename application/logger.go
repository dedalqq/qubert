package application

import (
	"io"
	"os"
	"sync"
	"time"
	"unsafe"
)

const (
	dateOffset        = 0
	FirstHeaderOffset = 32
)

type systemLogger struct {
	created time.Time
	mx      sync.Mutex
	file    *os.File
}

const (
	typeFlagMasq = 0x80
)

func isChainHeader(data []byte) bool {
	return data[0]&typeFlagMasq == 0
}

func isMessage(data []byte) bool {
	return data[0]&typeFlagMasq != 0
}

func openLogStorage(fileName string) (*systemLogger, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var created time.Time

	if info.Size() == 0 {
		created, err = initStorage(f)
	} else {
		created, err = checkAndReadCreated(f)
	}

	if err != nil {
		return nil, err
	}

	return &systemLogger{
		created: created,
		file:    f,
	}, nil
}

type chain struct {
}

func fileSize(f *os.File) int64 {
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}

	return info.Size()
}

func storageIsEmpty(f *os.File) bool {
	return fileSize(f) == 32
}

func writeBlockData(data []byte, flag byte, a uint16, b uint32, c uint32, d uint32) {
	data[0] = flag
	data[1] = 0x00

	dd16 := *(*[2]byte)(unsafe.Pointer(&a))

	data[2] = dd16[0]
	data[3] = dd16[1]

	dd32 := *(*[4]byte)(unsafe.Pointer(&b))

	data[4] = dd32[0]
	data[5] = dd32[1]
	data[6] = dd32[2]
	data[7] = dd32[3]

	dd32 = *(*[4]byte)(unsafe.Pointer(&c))

	data[8] = dd32[0]
	data[9] = dd32[1]
	data[10] = dd32[2]
	data[11] = dd32[3]

	dd32 = *(*[4]byte)(unsafe.Pointer(&d))

	data[12] = dd32[0]
	data[13] = dd32[1]
	data[14] = dd32[2]
	data[15] = dd32[3]

}

func writeChainHeader(f *os.File, id string) error {
	_, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	data := make([]byte, 16)
	writeBlockData(data, 0x00, uint16(len(id)), 0, 0, 0)

	_, err = f.WriteAt(data, FirstHeaderOffset)
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(id))
	if err != nil {
		return err
	}

	return nil
}

func (l *systemLogger) chain(id string) (*chain, error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	var err error

	if storageIsEmpty(l.file) {
		err = writeChainHeader(l.file, id)
		if err != nil {
			return nil, err
		}

		return &chain{}, nil
	}

	data := make([]byte, 16)

	_, err = l.file.ReadAt(data, FirstHeaderOffset)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func initStorage(f *os.File) (time.Time, error) {
	created := time.Now()

	binData, err := created.GobEncode()
	if err != nil {
		panic(err)
	}

	data := make([]byte, 32)
	_, err = f.WriteAt(data, 0)
	if err != nil {
		return time.Time{}, err
	}

	_, err = f.WriteAt(binData, dateOffset)
	if err != nil {
		return time.Time{}, err
	}

	return created, f.Sync()
}

func checkAndReadCreated(f *os.File) (time.Time, error) {
	data := make([]byte, 32)
	_, err := f.ReadAt(data, dateOffset)
	if err != nil {
		return time.Time{}, err
	}

	created := time.Time{}

	err = created.GobDecode(data)
	if err != nil {
		panic(err)
	}

	return created, nil
}
