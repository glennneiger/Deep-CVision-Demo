package main

import (
	"github.com/gorilla/websocket"
	"github.com/ddliu/go-httpclient"
	"net/http"
	//"os"
	"fmt"
	"log"
	"regexp"
	"strconv"
	//"log"
	//"io/ioutil"
//	"time"
	"encoding/json"
	"strings"
	"bytes"
	"errors"
	"bufio"
	"os/exec"
)


var dn_is_running bool = false
var gs_is_running bool = false
var fps int = 6
var use_png bool = true

type BambStream struct {
	Result []struct {
		Vid string `json:"vid"`
		ID string `json:"id"`
		Title string `json:"title"`
		Type string `json:"type"`
		Username string `json:"username"`
		Author string `json:"author"`
		Created string `json:"created"`
		//Length int `json:"length"`
		Replaced string `json:"replaced"`
		Lat string `json:"lat"`
		Lon string `json:"lon"`
		Country string `json:"country"`
		Trail string `json:"trail"`
		Upvotes string `json:"upvotes"`
		Preview string `json:"preview"`
		URL string `json:"url"`
		Mirrors string `json:"mirrors"`
		AccessMode string `json:"access_mode"`
		Owner struct {
			Name string `json:"name"`
			UID int `json:"uid"`
			MostRecent string `json:"mostRecent"`
			Unlisted bool `json:"unlisted"`
			Hidden bool `json:"hidden"`
			Timezone int `json:"timezone"`
			TimezoneNameShort string `json:"timezone_name_short"`
			Avatar struct {
				Small struct {
					Filename string `json:"filename"`
					Size int `json:"size"`
				} `json:"small"`
				Large struct {
					Filename string `json:"filename"`
					Size int `json:"size"`
				} `json:"large"`
			} `json:"avatar"`
			ProfileURL string `json:"profile_url"`
		} `json:"owner"`
		//ViewsLive json.RawMessage `json:"views_live"`
		//ViewsTotal json.RawMessage `json:"views_total"`
		Thumbnails struct {
			One20X80 struct {
				Default string `json:"default"`
			} `json:"120x80"`
			One40X115 struct {
				Default string `json:"default"`
			} `json:"140x115"`
			One60X120 struct {
				Default string `json:"default"`
			} `json:"160x120"`
			Three20X240 struct {
				Default string `json:"default"`
			} `json:"320x240"`
		} `json:"thumbnails"`
		PreviewRaw string `json:"preview_raw"`
		PreviewThumbnail string `json:"preview_thumbnail"`
		PreviewThumbnail120X80 string `json:"preview_thumbnail_120x80"`
		CommentCount int `json:"comment_count"`
		Page string `json:"page"`
		Framerate string `json:"framerate"`
		Visibility string `json:"visibility"`
		ChatLoginRequired bool `json:"chat_login_required"`
		Tags []interface{} `json:"tags"`
		DeviceName string `json:"device_name"`
		DeviceClass string `json:"device_class"`
		PositionAccuracy string `json:"position_accuracy"`
		PositionType string `json:"position_type"`
		Width string `json:"width"`
		Height string `json:"height"`
	} `json:"result"`
}

type Person struct {
    Name string
    Age  int
}

var upgrader = websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
}

const (
	USERAGENT       = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"
	TIMEOUT         = 30
	CONNECT_TIMEOUT = 5
	SERVER          = "https://bambuser.com/xhr-api/index.php?username=r00tz&sort=live&access_mode=0%2C1%2C2&limit=1&_strict=1&method=broadcast&format=json&_=1497954399229"
)

func get_bambuser_meta() (*BambStream,error) {

	//var url string

	httpclient.Defaults(httpclient.Map{
		"opt_useragent":   USERAGENT,
		"opt_timeout":     TIMEOUT,
		"Accept-Encoding": "gzip, deflate, sdch, br",
	})	

	res, _ := httpclient.
		WithHeader("Accept-Language", "sv-SE,sv;q=0.8,en-US;q=0.6,en;q=0.4").
		WithHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		WithHeader("Accept", "application/json, text/javascript, */*; q=0.01").
		WithHeader("Referer", "https://bambuser.com/channel/r00tz").
		WithHeader("X-Requested-With", "XMLHttpRequest").
		WithHeader("Connection", "keep-alive").
		WithHeader("Cookie","_hjIncludedInSample=1; fbm_19126885616=base_domain=.bambuser.com; __utmt=1; SESS28188803e232eb80b00ec66d12f3d957=g5fbmpmsemhdqgjtbqift65696; active_session=1; fbsr_19126885616=ipy4wBwDuNonOyvpjJh-74SrSwoNZzEPpYI0S3g6Hbo.eyJhbGdvcml0aG0iOiJITUFDLVNIQTI1NiIsImNvZGUiOiJBUURDODBaTG5qLXJRcFJtaVZQNF9VMXlxQ2JPNU8zcXljZEZycEpocXJUa0haUVI2enNLV3R6VXdpZ3U2eUo1WmtObDEzQ2FhaGI1MXYwNVVYM3R1VEx1QnktbUZoZjlzVy1YdlZqZnJGMzlWNWpFMmN4VTQ3VmVoNTcxVjhqd0hNaXRqV3EwbzNNT0JqNmtXOHlJcVFIYlY4ZVdqVkNmV3NtRlZPMTdnZElobVB1MlZORXNfUEZWcnFvUTc4OXh2Zl9hWFVMWEx0OHYyZnVIckRBa00xSjhabzJoQXhVNGl4SGgzRW80X2p1Q0dIdTNBNnYwLTJ2WGVGVG1lc3ZvNy1yNjFFN2R3aFlobGx0emxuaVd1N2I5TWFWY1ZVelFiVmI0SXV4SnhFUHhkRG1YLUpwd04zRU03UUJyaVNoS1daRnVkUk1mNjNNZFJYRWFQQm4yUktkWSIsImlzc3VlZF9hdCI6MTQ5Nzk1NDI5OCwidXNlcl9pZCI6IjEwMTU0NDA5NDYwODY2MDg2In0; b_fb_token=refresh; __utma=27127202.652164678.1497653558.1497949541.1497952506.3; __utmb=27127202.10.10.1497952506; __utmc=27127202; __utmz=27127202.1497949541.2.2.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided)").
		Get(SERVER, nil)

	str,_ := res.ToString()
	str = strings.TrimRight(str,string(0))
	//fmt.Println(str)
	streams := &BambStream{};
	//err := json.Unmarshal([]byte(res.ToString), &streams)
	fmt.Printf("%s\n",str)
	err := json.Unmarshal(bytes.Trim(([]byte(str)),"\x00"), &streams)
	if err != nil {
		return streams,err
   	}
	//fmt.Printf("Type: %s\n", streams.Result[0].Type)
	//fmt.Printf("Url: %s\n", streams.Result[0].URL)
	//fmt.Printf("type: %s\n",streams.Result[0].Type)
	if len(streams.Result) > 0 && streams.Result[0].Type == "live" {

	} else {
		err = errors.New("No live stream found")
	}
	return streams,err
}

func get_stream_url()(string) {

		out, _ := exec.Command("sh","-c","/home/ubuntu/websocklab/bambscrape.pl").Output()
		return strings.TrimRight(string(out), "\n")
}

func google_speech_streamer(url string, msg chan string) {

	//msg := make(chan string)
	//errors := make(chan error, 0)
	go func(){


		//cmdName := "./outputter.pl "+url
		ffmpeg := "ffmpeg -i "+url+" -loglevel quiet  -f s16le -ar 16000 -ac 1 -vn -acodec pcm_s16le -"
		gspeech := "go run /home/ubuntu/golang-samples/speech/livecaption/livecaption.go"
		fmt.Printf("%s\n",ffmpeg)
		fmt.Printf("%s\n",gspeech)

		ffmpegArgs := strings.Fields(ffmpeg)
		gspeechArgs := strings.Fields(gspeech)

		ffmpegcmd := exec.Command(ffmpegArgs[0], ffmpegArgs[1:len(ffmpegArgs)]...)
		gspeechcmd := exec.Command(gspeechArgs[0], gspeechArgs[1:len(gspeechArgs)]...)
		//gspeechcmd.Dir = "/home/ubuntu/gspeech"

		gspeechcmd.Stdin, _ = ffmpegcmd.StdoutPipe()
		//gspeechcmd.Stdout = os.Stdout
		gspeechout, _ := gspeechcmd.StdoutPipe()
		reader := bufio.NewReader(gspeechout)
		var err error;
		fmt.Printf("gspeechcmd.Start\n")
		if err = gspeechcmd.Start(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ffmpegcmd.Run\n")
		if err = ffmpegcmd.Start(); err != nil {
			log.Fatal(err)
		}
		/*
		if err = gspeechcmd.Wait(); err != nil {
			log.Fatal(err)
		}
		*/

		for {
			//fmt.Printf("waiting for data on reader\n")
			line, err := reader.ReadString('\n')
			//fmt.Printf("stdout > Read %d characters\n", len(line))
			if err != nil {
				fmt.Println(line)
				if err := ffmpegcmd.Process.Kill(); err != nil {
					log.Fatal("failed to kill: ", err)
				}
				msg <- "gEOF"
				//log.Fatal(err)
				return;
			}
			line = strings.TrimRight(line, "\n")
			line = strings.Replace(line,"\\303\\244","ä",-1)
			line = strings.Replace(line,"\\303\\245","å",-1)
			line = strings.Replace(line,"\\303\\266","ö",-1)
			line = strings.Replace(line,"\\303\\205","Å",-1)
			line = strings.Replace(line,"\\303\\204","Ä",-1)
			line = strings.Replace(line,"\\303\\226","Ö",-1)
			fmt.Printf("%s\n",line);
			r := regexp.MustCompile(`transcript:"(?P<caption>[^"]*)"(?: confidence:(?P<confidence>[0-9.]+))?`)
			m := r.FindStringSubmatch(line);
			if len(m) == 3 {
				confidence,_ := strconv.ParseFloat(m[2],64);
				msg <- fmt.Sprintf(`{"caption":"%s","confidence":%f}`,m[1],confidence)
		    } else if len(m) == 2 {
				msg <- fmt.Sprintf(`{"caption":"%s"}`,m[1])
			}
			fmt.Printf("%#v\n", m)
		}
	}()
	return
}

func init_darknet() *exec.Cmd {
		fmt.Printf("Initializing darknet...")
		png := func()string{if use_png { return "-png" } else { return "" }}()
		darknet := "./darknet detector tndemo "+png+" cfg/coco.data cfg/yolo.cfg yolo.weights 2> /tmp/service.stderr"
		darknetArgs := strings.Fields(darknet)

		darknetcmd := exec.Command(darknetArgs[0], darknetArgs[1:len(darknetArgs)]...)
		darknetcmd.Dir = "/home/ubuntu/darknet"
		var err error;
		fmt.Printf("darknetcmd.Start\n")
		if err = darknetcmd.Start(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("done!\n");

		return darknetcmd;
}

func ffmpeg_streamer(url string, msg chan string, darknetcmd *exec.Cmd) {

	//msg := make(chan string)
	//errors := make(chan error, 0)
	go func(){


		//cmdName := "./outputter.pl "+url
		ffmpeg := "ffmpeg -i "+url+" -loglevel quiet -an -f rawvideo -r "+strconv.Itoa(fps)+" -pix_fmt rgb24 -"
		fmt.Printf("%s\n",ffmpeg)

		ffmpegArgs := strings.Fields(ffmpeg)
		ffmpegcmd := exec.Command(ffmpegArgs[0], ffmpegArgs[1:len(ffmpegArgs)]...)

		darknetcmd.Stdin, _ = ffmpegcmd.StdoutPipe()
		//darknetcmd.Stdout = os.Stdout
		darknetout, _ := darknetcmd.StdoutPipe()
		reader := bufio.NewReader(darknetout)
		var err error;
		/*
		fmt.Printf("darknetcmd.Start\n")
		if err = darknetcmd.Start(); err != nil {
			log.Fatal(err)
		}
		*/
		fmt.Printf("ffmpegcmd.Run\n")
		if err = ffmpegcmd.Start(); err != nil {
			log.Fatal(err)
		}
		/*
		if err = darknetcmd.Wait(); err != nil {
			log.Fatal(err)
		}
		*/

		var json string
		for {
			//fmt.Printf("waiting for data on reader\n")
			line, err := reader.ReadString('\n')
			//fmt.Printf("stdout > Read %d characters\n", len(line))
			if err != nil {
				fmt.Println(line)
				//log.Fatal(err)
				msg <- "EOF"
				return;
			}
			line = strings.TrimRight(line, "\n")
			//fmt.Printf("%s\n",line);
			json = json+line
			if len(line) >= 2 && line[len(line)-2:] == "]}" {
				fmt.Printf("json message length: %d\n",len(json))
				msg <- json
				json = ""
			}
	
			//msg <- line
		}
	}()
	return
}

func darknet_streamer(url string, msg chan string) {

	//msg := make(chan string)
	//errors := make(chan error, 0)
	go func(){


		//cmdName := "./outputter.pl "+url
		//png := func()string{if use_png { return "-png" } else { return "" }}()
		ffmpeg := "ffmpeg -i "+url+" -loglevel quiet -an -f rawvideo -r "+strconv.Itoa(fps)+" -pix_fmt rgb24 -"
		//darknet := "./darknet detector tndemo "+png+" cfg/coco.data cfg/yolo.cfg yolo.weights 2> /tmp/service.stderr"
		darknet := "./darknet detect cfg/yolo.cfg yolo.weights ../rl2.png"
		//darknet := "./darknet detector tndemo -hier .2 "+png+" cfg/combine9k.data cfg/yolo9000.cfg yolo9000.weights 2> /tmp/service.stderr"
		fmt.Printf("%s\n",ffmpeg)
		fmt.Printf("%s\n",darknet)

		ffmpegArgs := strings.Fields(ffmpeg)
		darknetArgs := strings.Fields(darknet)

		ffmpegcmd := exec.Command(ffmpegArgs[0], ffmpegArgs[1:len(ffmpegArgs)]...)
		darknetcmd := exec.Command(darknetArgs[0], darknetArgs[1:len(darknetArgs)]...)
		darknetcmd.Dir = "/home/ubuntu/new_darknet/darknet"

		darknetcmd.Stdin, _ = ffmpegcmd.StdoutPipe()
		//darknetcmd.Stdout = os.Stdout
		darknetout, _ := darknetcmd.StdoutPipe()
		reader := bufio.NewReader(darknetout)
		var err error;
		fmt.Printf("darknetcmd.Start\n")
		if err = darknetcmd.Start(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ffmpegcmd.Run\n")
		if err = ffmpegcmd.Start(); err != nil {
			log.Fatal(err)
		}
		/*
		if err = darknetcmd.Wait(); err != nil {
			log.Fatal(err)
		}
		*/

		var json string
		for {
			//fmt.Printf("waiting for data on reader\n")
			line, err := reader.ReadString('\n')
			//fmt.Printf("stdout > Read %d characters\n", len(line))
			if err != nil {
				fmt.Println(line)
				//log.Fatal(err)
				msg <- "EOF"
				return;
			}
			//fmt.Println(line)
			line = strings.TrimRight(line, "\n")
			//fmt.Printf("%s\n",line);
			json = json+line
			if len(line) >= 2 && line[len(line)-2:] == "]}" {
				fmt.Printf("json message length: %d\n",len(json))
				msg <- json
				json = ""
			}
	
			//msg <- line
		}
	}()
	return
}


func main() {

	/*
    indexFile, err := os.Open("html/index.html")
    if err != nil {
        fmt.Println(err)
    }
    index, err := ioutil.ReadAll(indexFile)
    if err != nil {
        fmt.Println(err)
    }
	*/

    var chans []chan string 
    dch := make(chan string)

	//darknetcmd := init_darknet();

    http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {

		this_ch :=  make(chan string)
		chans = append(chans, this_ch)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Client subscribed")
		/*
		bamb,err := get_bambuser_meta()
		if err != nil {
			fmt.Println(err)
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"error\":\"%s\"}",err)))
			return
		}
		//err = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"bambref\":\"%s\"}",url)))
		js, err := json.Marshal(bamb);
		err = conn.WriteMessage(websocket.TextMessage, []byte(js))
		*/
		url := get_stream_url();
		if url == "" {
			fmt.Println("No stream found")
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"error\":\"%s\"}","No stream found")))
			return;
		} 
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"result\":[{\"url\":\"%s\"}]}",url)))
		if !dn_is_running {
			dn_is_running = true
			darknet_streamer(url,dch)
			//ffmpeg_streamer( bamb.Result[0].URL, dch, darknetcmd )
		}
		if !gs_is_running {
			gs_is_running = true
			google_speech_streamer(url,dch)
		}


		for {
		  //time.Sleep(2 * time.Second)
		  if gs_is_running == false {
			  gs_is_running = true
			  google_speech_streamer(url,dch)
			}
		  if dch != nil {
			out := <- dch
			if out == "EOF" {
				dn_is_running = false
				conn.Close()
				break;
			}
			if out == "gEOF" {
				gs_is_running = false
				fmt.Printf("Google Speech disconnected!\n");
				err = conn.WriteMessage(websocket.TextMessage, []byte("{\"error\":\"Google Speech disconnected\"}"))
				/*
				bamb,err = get_bambuser_meta()
				if err != nil {
					fmt.Println(err)
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"error\":\"%s\"}",err)))
					return
				}
				*/
				url := get_stream_url();
				if url == "" {
					fmt.Println("No stream found")
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"error\":\"%s\"}","No stream found")))
					return;
				} 
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"result:[{\"url\":\"%s\"}]}",url)))
				continue;
			}
			fmt.Printf("Sending on websocket, length: %d\n",len(out))
			err = conn.WriteMessage(websocket.TextMessage, []byte(out))
			if err != nil {
			  fmt.Println(err)
			  break
			}
		  } else {
			conn.Close()
			break
		  }
		}
		fmt.Println("Client unsubscribed")
    })


	/*
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, string(index))
    })
	*/

    http.ListenAndServe(":3000", nil)
}
