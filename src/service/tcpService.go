package service

import (
	"commitor"
	"dnf"

	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func tcpIndexHandler(conn net.Conn) {
	/*
	   protocol:
	   struct pkg {
	       int32_t magic; // 0xDEAD
	       int32_t size;  // LITTLE ENDIAN
	       char    data[];
	   };
	*/
	defer conn.Close()

	for {
		var magic, size int32
		var data []byte

		if err := binary.Read(conn, binary.LittleEndian, &magic); err != nil {
			fmt.Println("Read magic error:", err)
			return
		}

		if magic != 0xDEAD {
			fmt.Println("magic number error:", magic)
			return
		}

		if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
			fmt.Println("Read size error:", err)
			return
		}

		if size <= 0 {
			fmt.Println("Size error:", size)
			return
		}

		data = make([]byte, int(size))

		if _, err := io.ReadFull(conn, data); err != nil {
			fmt.Println("Read data error:", err)
			return
		}

		if err := handleIndexRequestData(conn, data); err != nil {
			fmt.Println("handleIndexRequestData error:", err)
			return
		}
	}
}

func handleIndexRequestData(conn net.Conn, data []byte) error {
	conds := make([]dnf.Cond, 0)
	params := strings.Split(string(data), "&")
	if len(params) == 0 {
		return errors.New("param error")
	}
	for i := 0; i != len(params); i++ {
		kv := strings.SplitN(params[i], "=", 2)
		if len(kv) != 2 {
			continue
		}

		if kv[0] == "query" {
			kv[1], _ = url.QueryUnescape(kv[1])
			vals := strings.Split(kv[1], "/")
			if len(vals) <= 1 {
				continue
			}

			switch {
			case vals[1] == "0": /* splash: /0/[width]/[height] */
				if len(vals) != 4 {
					fmt.Println("query string err: ", kv[1])
					continue
				}
				conds = append(conds, dnf.Cond{Key: "width", Val: vals[2]})
				conds = append(conds, dnf.Cond{Key: "height", Val: vals[3]})

			case vals[1] == "1" || vals[1] == "3":
				/* banner: /1/[cateid]/[pos]/[width]/[height] */
				/* recommend: /3/[cateid]/[pos]/[width]/[height] */
				if len(vals) != 6 {
					fmt.Println("query string err: ", kv[1])
					continue
				}
				conds = append(conds, dnf.Cond{Key: "width", Val: vals[4]})
				conds = append(conds, dnf.Cond{Key: "height", Val: vals[5]})

			default:
				continue
			}

		} else {
			//fmt.Println("cond: ", kv[0], kv[1])
			conds = append(conds, dnf.Cond{Key: kv[0], Val: kv[1]})
		}
	}

	var repData string

	//fmt.Printf("search cond: %+v\n", conds)

	h := dnf.GetHandler()
	if h == nil {
		return errors.New("cannot get dnf handler")
	}

	now := time.Now()

	if docs, err := h.Search(conds); err != nil {
		fmt.Println(now.Format("2006-01-02 15:04:05"), "SearchErr:", err)
		return err
	} else {
		adlist := make([]interface{}, 0)
		for _, doc := range docs {
			if m := dnf.DocId2Map(doc); m != nil {
				adlist = append(adlist, dnf.DocId2Map(doc))
			}
		}
		m := make(map[string]interface{})
		m["data"] = adlist
		rc, _ := json.Marshal(m)
		repData = string(rc)
		fmt.Println(now.Format("2006-01-02 15:04:05"), "SearchOk:", repData)
	}

	magic := int32(0xBEEF)
	size := int32(len(repData))
	binary.Write(conn, binary.LittleEndian, &magic)
	binary.Write(conn, binary.LittleEndian, &size)
	_, err := conn.Write([]byte(repData))
	return err
}

func tcpZonesHandler(conn net.Conn) {
	/*
	   protocol:
	   struct pkg {
	       int32_t magic; // 0xCAFE
	       int32_t size;  // LITTLE ENDIAN
	       int32_t version; // version == 5 or version == 6
	       char    data[];
	   };
	*/
	defer conn.Close()

	for {
		var magic, size, version int32
		var data []byte

		if err := binary.Read(conn, binary.LittleEndian, &magic); err != nil {
			fmt.Println("Read magic error:", err)
			return
		}

		if magic != 0xCAFE {
			fmt.Println("magic number error:", magic)
			return
		}

		if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
			fmt.Println("Read size error:", err)
			return
		}

		if size <= 0 {
			fmt.Println("Size error:", size)
			return
		}

		if err := binary.Read(conn, binary.LittleEndian, &version); err != nil {
			fmt.Println("Read version error:", err)
			return
		}

		data = make([]byte, int(size))

		if _, err := io.ReadFull(conn, data); err != nil {
			fmt.Println("Read data error:", err)
			return
		}

		handleZoneRequestData(conn, data, int(version))
	}
}

func handleZoneRequestData(conn net.Conn, data []byte, version int) {
	m := make(map[string]interface{})

	h := commitor.GetZonesInfoHandler()
	zones := h.GetZonesInfo(version)

	poslist := make([]interface{}, 0)
	for i := 0; i != len(zones); i++ {
		item := make(map[string]interface{})
		item["posid"] = zones[i].Zoneid
		item["posdesc"] = zones[i].Comments
		if zones[i].Width != "-3" && zones[i].Height != "-3" {
			item["posquery"] = zones[i].Comments + "/" + zones[i].Width + "/" + zones[i].Height
		} else {
			item["posquery"] = zones[i].Comments
		}
		poslist = append(poslist, item)
	}

	m["data"] = poslist
	rc, _ := json.Marshal(m)

	magic := int32(0xFEED)
	repData := string(rc)
	size := int32(len(repData))
	binary.Write(conn, binary.LittleEndian, &magic)
	binary.Write(conn, binary.LittleEndian, &size)
	version32 := int32(version)
	binary.Write(conn, binary.LittleEndian, &version32)
	conn.Write([]byte(repData))
	// fmt.Println("return data: ", repData)
}

func doTcpServe(port int, tcpHandler func(net.Conn)) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic("Listen tcp :" + strconv.Itoa(port) + "error" + err.Error())
	}
	defer ln.Close()
	for {
		if conn, err := ln.Accept(); err == nil {
			//fmt.Println("TcpServe Accept ok")
			go tcpHandler(conn)
		} else {
			fmt.Println("TcpServe Accept conn error:", err)
		}
	}
}

func TcpServe() {
	// TODO: read port from conf file
	go doTcpServe(7778, tcpZonesHandler)
	go doTcpServe(7779, tcpIndexHandler)
}
