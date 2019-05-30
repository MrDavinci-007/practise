package main

import (
	"fmt"
	"github.com/piex/transcode"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var js int =0

type bookInfo struct{
	title,atuhor,class,link,img string
	list map[string]string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()/2)
	//getResult,getErr:=httpGet("https://www.69shu.com/allvisit_5_0_0_0_1_0_10.htm")
	total:=2
	var wg sync.WaitGroup

	startTime := time.Now()
	for page:=1;page<=total ;page++  {
		wg.Add(1)
		url:=fmt.Sprintf("https://www.69shu.com/allvisit_5_0_0_0_1_0_%v.htm",page)
		go spiderPage(url,&wg)
	}
	wg.Wait()
	endTime := time.Now()
	fmt.Printf("Process time %s", endTime.Sub(startTime))
}

//采集页面中列表链接
func spiderPage(urls string,wg *sync.WaitGroup){
	log.Printf("[ %v ]：列表抓取开始...\n",urls)
	getResult,getErr:=httpGet(urls)

	if getErr==nil{
		log.Printf("[ %v ]：列表抓取成功...\n",urls)
		log.Printf("[ %v ]：分析列表数据...\n",urls)
		html:=regexp.MustCompile(`<a target="_blank" href="(?s:(.*?))"`).FindAllStringSubmatch(getResult,-1)
		length:=len(html)
		if length>1{
			//为接下来每个作品信息页创建一个通道，用于并发
			fmt.Println(length)
			wg.Add(length-1)
			for i,_:=range html{
				go spiderInfo(html[i][1],wg)
				//break
			}
		}else{
			log.Printf("[ %v ]：列表数据为空...\n",urls)
			wg.Done()
		}
	}else{
		log.Printf("[ %v ]：列表抓取失败...\n",urls)
		wg.Done()
	}

}

//获取作品信息和目录链接
func spiderInfo(url string,wg *sync.WaitGroup) {

	log.Printf("[ %v ]：开始获取作品信息...\n",url)
	body,err:=httpGet(url)
	if err==nil{
		title:=regexp.MustCompile(` title="(.*?)"></a>`).FindAllStringSubmatch(body,1)
		link:=regexp.MustCompile(`<a class="button read" href="(.*?)">`).FindAllStringSubmatch(body,1)
		img:=regexp.MustCompile(`src="/files(.*?)"`).FindAllStringSubmatch(body,1)
		author_class:=regexp.MustCompile(`target="_blank">(.*?)</a>`).FindAllStringSubmatch(body,2)
		info:=bookInfo{
			title:transcode.FromString(title[0][1]).Decode("GBK").ToString(),
			link:link[0][1],
			img:"https://www.69shu.com/files"+img[0][1],
			atuhor:transcode.FromString(author_class[0][1]).Decode("GBK").ToString(),
			class:transcode.FromString(author_class[1][1]).Decode("GBK").ToString(),
		}
		go spiderDir(link[0][1],&info,wg)
		//fmt.Println(info)
	}else{
		log.Printf("[ %v ]：作品信息获取失败...\n",url)
	}


}

//获取目录页面所有的链接地址
func spiderDir(url string,info *bookInfo,wg *sync.WaitGroup){
	log.Printf("[ %v ]：开始获取目录链接...\n",url)
	body,err:=httpGet(url)
	if err==nil{
		link:=regexp.MustCompile(`<ul class="mulu_list">(?s:(.*?))</ul>`).FindAllStringSubmatch(body,2)
		//fmt.Println(link[1][1])
		if len(link)>1{
			_url:=regexp.MustCompile(`<a href="(.*?)"`).FindAllStringSubmatch(link[1][1],-1)
			_title:=regexp.MustCompile(`">(.*?)</a></li>`).FindAllStringSubmatch(link[1][1],-1)

			_list:=make(map[string]string,len(_url))
			for i,_:=range _url{
				//fmt.Println(_url[i][1],transcode.FromString(_title[i][1]).Decode("GBK").ToString())
				_list[_url[i][1]]=transcode.FromString(_title[i][1]).Decode("GBK").ToString()
			}
			info.list=_list
			log.Printf("[ %v ]：目录链接获取成功...\n",url)
		}
	}else{
		log.Printf("[ %v ]：获取目录链接失败...\n",url)
	}
	js+=1
	fmt.Println("--counter---",strconv.Itoa(js))
	wg.Done()
}

func httpGet(url string)(data string,err error){

	resp,respErr:=http.Get(url)

	if(respErr==nil){
		body,bodyErr:=ioutil.ReadAll(resp.Body)
		if bodyErr!=nil{
			err=bodyErr
			log.Printf("[ %v ]：读取数据出错...\n",url)
		}
		data=string(body)
	}else{
		err=respErr
		log.Printf("[ %v ]：获取数据出错...\n",url)
	}
	//defer resp.Body.Close()
	return
}