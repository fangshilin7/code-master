// SPFDAnalyser project Analyser.go

/*
Analyser document
*/
package Analyser

import (
	"fmt"
	//"log"
	"net"
	"os"
	"sync"
	"time"
	"unsafe"

	"git.scsv.online/go/logger"
)

type Analyser struct {
	conn      net.Conn
	addr      string
	id        string
	valid_buf int
	buff      [1024 * 160]byte
	//buff [160]byte
}

type pkt_header struct {
	Pkt_header  [4]byte
	Pro_version [2]byte
	Pkt_type    [2]byte
	Pkt_len     [4]byte
	reserve     [3]byte
}

type req_body struct {
	Req_mode           byte
	Req_content        [2]byte
	Total_chnl_num     byte
	Effective_chnl_num byte
	Chnl_id            byte
	Effective_sign     byte
	reserve            byte
	time               [8]byte
}

type peo_num struct {
	Data_type             [2]byte
	Data_len              [4]byte
	Number                [4]byte
	Congestion_level      byte
	Congestion_level_best byte
	Car_num               byte
	reserve               [5]byte
}

type pic struct {
	Data_type [2]byte
	Data_len  [4]byte
}

var file_mutex sync.Mutex
var simu_mutex sync.Mutex

var simu_data []byte = make([]byte, 1024*160)

//var online_data []byte = make([]byte, 32)

//var read_buf [30]byte //20

func getCrc(buf []byte) byte {

	var crc uint16 = 0x0
	for i := 0; i < len(buf); i++ {
		crc = crc ^ (uint16)(buf[i])
		for j := 0; j < 8; j++ {
			if crc&0x80 > 0 {
				crc = (crc << 1) ^ 0x07
			} else {
				crc <<= 1
			}
		}
	}
	//logger.Info("getCrc 111 %d", crc)
	crc &= 0xff
	//logger.Info("getCrc 222 %d", crc)
	return (byte)(crc)
}

func Simulation_data_load() (int, error) {
	fmt.Println("Simulation_data_load start")
	path, _ := os.Getwd()
	pkt_pic_start := 16 + 16 + 18 + 6
	_, pic_len, err := file2Bytes(pkt_pic_start, path+"/test.jpg")
	if err != nil {
		fmt.Println("file2Bytes err", err)
		return 0, err
	}

	pkt_len := pkt_pic_start + pic_len
	fmt.Println("len:", pic_len, pkt_len)

	//data := make([]byte, pkt_len+16)

	header, _ := bulid_header(pkt_len-16, 0x03)
	idex_start := 0
	idex_end := 16
	copy(simu_data[idex_start:idex_end], header)

	fmt.Println("Simulation_data_load start1")
	body_info, _ := bulid_body_info()
	idex_start += idex_end - idex_start
	idex_end = idex_start + 16
	copy(simu_data[idex_start:idex_end], body_info)

	fmt.Println("Simulation_data_load start2")
	peo_info, _ := bulid_peo_info()
	idex_start += idex_end - idex_start
	idex_end = idex_start + 18
	copy(simu_data[idex_start:idex_end], peo_info)

	fmt.Println("Simulation_data_load start3")
	pic_info, _ := bulid_pic_info(pic_len)
	idex_start += idex_end - idex_start
	idex_end = idex_start + 6
	copy(simu_data[idex_start:idex_end], pic_info)

	fmt.Println("Simulation_data_load end")
	return pkt_len, nil
}

func (analyser *Analyser) Start(addr string, id string, pkt_len int) error {
	analyser.id = id
	analyser.valid_buf = pkt_len
	analyser.addr = addr

	fmt.Println("Analyser Start 1")
	//var comm_buf []byte = make([]byte, pkt_len)
	simu_mutex.Lock()
	copy(analyser.buff[:], simu_data[:pkt_len])
	simu_mutex.Unlock()

	fmt.Println("Analyser Start 2")
	for {
		err := analyser.tcpConn()
		if err != nil {
			logger.Error("net DialTcp conn[5] id[%s] err[%v]", id, err)
			//fmt.Println("net DialTcp conn[5] err", err)
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}
		// tConn, ok := conn.(*net.TCPConn)
		// if !ok {
		// 	fmt.Println("vsdfgfdghdfhgdfh")
		// }
		// tConn.SetWriteBuffer(1024 * 1024 * 10)
		//tConn.Write([]byte("test"))
		err = analyser.sendOnline()
		if err != nil {
			fmt.Println("sendOnline err", err)
			return err
		}
		break
	}
	logger.Info("tcp connect success id[%s]", id)
	fmt.Println("Analyser Start 3")
	go analyser.Run()
	return nil
}

func (analyser *Analyser) recvFixedData(conn net.Conn, size int) (recvBuf []byte, err error) {
	buf := make([]byte, size)
	sum := 0
	for {
		if sum == size {
			break
		}
		n, err := conn.Read(buf[sum:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
			return nil, err
		}
		sum += n
	}
	return buf, nil
}

func (analyser *Analyser) Run() {
	fmt.Println("Analyser run start")
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Fatalf("Analyser error:%v", err)
	// 		return
	// 	}
	// }()

	//defer analyser.conn.Close()
	read_err_count := 0
	var read_buf [20]byte //20
	for {
		// defer func() {
		// 	if err := recover(); err != nil {
		// 		log.Fatalf("loop error:%v", err)
		// 		return
		// 	}
		// }()

		fmt.Println("tcp Read start")
		//conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
		//_, err := analyser.recvFixedData(analyser.conn, 20)
		_, err := analyser.conn.Read(read_buf[:])
		if err != nil {
			fmt.Println("conn read err", err)
			read_err_count++
			if read_err_count > 10 {
				analyser.conn.Close()
				for {
					err = analyser.tcpConn()
					if err != nil {
						logger.Error("net DialTcp reconn[5] id[%s] err", analyser.id, err)
						//fmt.Println("net DialTcp reconn[5] err", err)
						time.Sleep(time.Duration(5) * time.Second)
						continue
					}
					break
				}

				read_err_count = 0
			}
			continue
		}
		//analyser.recvFixedData(analyser.conn, 20)
		fmt.Println("tcp Read success:", time.Now(), analyser.id)
		go analyser.sendResponse()
		// if err != nil {
		// 	fmt.Println("sendResponse err", err)
		// 	return
		// }
		//time.Sleep(time.Duration(10) * time.Second)
		//<-time.After(time.Second * time.Duration(10))
	}

	fmt.Println("Analyser run end")
}

func (analyser *Analyser) tcpConn() error {
	//tcpaddr, err := net.ResolveTCPAddr("", addr)
	var err error
	for con_count := 5; con_count > 0; con_count-- {
		//*conn, err = net.DialTCP("tcp4", nil, tcpaddr)
		analyser.conn, err = net.Dial("tcp", analyser.addr)
		if err != nil {
			fmt.Printf("net DialTcp[%s] err[%v]\n", analyser.addr, analyser.id, err.Error())
			continue
		}
		return nil
	}

	return err
}

func (analyser *Analyser) sendOnline() error {
	var online [32]byte
	header, _ := bulid_header(16, 0x01)
	copy(online[0:16], header)

	copy(online[16:29], []byte(analyser.id))
	fmt.Println("sendOnline data:", online)

	//conn.SetWriteDeadline(time.Now().Add(time.Duration(10) * time.Second))
	_, err := analyser.conn.Write(online[:])
	if err != nil {
		fmt.Println("sendOnline Write err", err)
		return err
	}
	fmt.Println("sendOnline end")
	return nil
}

func (analyser *Analyser) sendResponse() error {
	fmt.Println("sendResponse start", analyser.id)

	//conn.SetWriteDeadline(time.Now().Add(time.Duration(5) * time.Second))
	n, err := analyser.conn.Write(analyser.buff[:analyser.valid_buf])
	if err != nil {
		fmt.Println("sendResponse Write err", err)
		return err
	}
	fmt.Println("sendResponse end ", analyser.id, "size", n)
	return nil
}

func bulid_header(pkt_len int, pkt_type byte) ([]byte, error) {
	data := make([]byte, 16)
	fmt.Println("pkt_len:", pkt_len)
	header := &pkt_header{}

	header.Pkt_header[0] = 0x02
	header.Pkt_header[1] = 0x02
	header.Pkt_header[2] = 0x02
	header.Pkt_header[3] = 0x02

	header.Pro_version[1] = 0x14
	header.Pkt_type[1] = pkt_type

	_ = int2Bytes(&header.Pkt_len, pkt_len)

	p := unsafe.Pointer(&header)
	q := (*[]byte)(p)
	copy(data[0:15], (*q)[0:])

	crc := getCrc([]byte(data[0:15]))
	fmt.Println("crc:", crc)
	data[15] = crc
	fmt.Println("header data:", data)

	return data, nil
}

func bulid_body_info() ([]byte, error) {
	data := make([]byte, 16)
	body_info := &req_body{}

	body_info.Req_mode = 0x02
	body_info.Req_content[1] = 0x03
	body_info.Total_chnl_num = 0x01
	body_info.Effective_chnl_num = 0x01
	body_info.Chnl_id = 0x01
	body_info.Effective_sign = 0x01

	base_time := time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	body_time := time.Now().Unix() - base_time

	_ = int642Bytes(&body_info.time, body_time)

	p := unsafe.Pointer(&body_info)
	q := (*[]byte)(p)
	copy(data[0:16], (*q)[0:])

	fmt.Println("body_info data:", data, body_time)

	return data, nil
}

func bulid_peo_info() ([]byte, error) {
	data := make([]byte, 18)
	peo_info := &peo_num{}

	peo_info.Data_type[1] = 0x01
	Data_len := 12
	_ = int2Bytes(&peo_info.Data_len, Data_len)

	Number := 20
	_ = int2Bytes(&peo_info.Number, Number)

	peo_info.Congestion_level = 0x02
	peo_info.Congestion_level_best = 0x04
	peo_info.Car_num = 0x00

	p := unsafe.Pointer(&peo_info)
	q := (*[]byte)(p)
	copy(data[0:18], (*q)[0:])

	fmt.Println("peo_info data:", data)

	return data, nil
}

func bulid_pic_info(pic_len int) ([]byte, error) {
	data := make([]byte, 6)
	pic_info := &pic{}

	pic_info.Data_type[1] = 0x02

	_ = int2Bytes(&pic_info.Data_len, pic_len)

	p := unsafe.Pointer(&pic_info)
	q := (*[]byte)(p)
	copy(data[0:6], (*q)[0:])

	fmt.Println("pic_info data:", data)

	return data, nil
}

// 读取文件到[]byte中
func file2Bytes(pkt_pic_start int, filename string) ([]byte, int, error) {
	fmt.Println("file2Bytes start")
	// File
	file_mutex.Lock()
	defer file_mutex.Unlock()
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("file open err:", err)
		//file.Close()
		return nil, 0, err
	}
	defer file.Close()
	fmt.Println("file2Bytes start1")

	// FileInfo:
	stats, err := file.Stat()
	if err != nil {
		fmt.Println("file Stat err:", err)
		return nil, 0, err
	}
	pkt_pic_end := int(stats.Size()) + pkt_pic_start
	fmt.Println("file2Bytes start2", pkt_pic_start, stats.Size(), pkt_pic_end)
	// []byte
	//data := make([]byte, stats.Size())
	//fmt.Println("file2Bytes start3", stats.Size())
	count, err := file.Read(simu_data[pkt_pic_start:pkt_pic_end])
	if err != nil {
		fmt.Println("file Read err:", err)
		return nil, 0, err
	}
	fmt.Printf("read file %s len: %d \n", filename, count)
	return nil, count, nil
}

func int2Bytes(data *[4]byte, value int) error {
	(*data)[0] = byte((value >> 24) & 0xff)
	(*data)[1] = byte((value >> 16) & 0xff)
	(*data)[2] = byte((value >> 8) & 0xff)
	(*data)[3] = byte(value & 0xff)
	return nil
}

func int642Bytes(data *[8]byte, value int64) error {
	(*data)[0] = byte((value >> 56) & 0xff)
	(*data)[1] = byte((value >> 48) & 0xff)
	(*data)[2] = byte((value >> 40) & 0xff)
	(*data)[3] = byte((value >> 32) & 0xff)
	(*data)[4] = byte((value >> 24) & 0xff)
	(*data)[5] = byte((value >> 16) & 0xff)
	(*data)[6] = byte((value >> 8) & 0xff)
	(*data)[7] = byte(value & 0xff)
	return nil
}

func struct2Bytes(from interface{}, len int) ([]byte, error) {
	data := make([]byte, len)

	p := unsafe.Pointer(&from)
	q := (*[]byte)(p)
	copy(data[0:], (*q)[0:])

	return data, nil
}
