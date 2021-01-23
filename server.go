package main

import (
	"fmt"
	"os"
	"sync"
	"net"
	"encoding/binary"
)

/*
一款石头剪刀布的小游戏
游戏介绍：游戏一共进行10轮，每轮中胜者获得2分，败者获取0分，如果平局则各获得一分
主要涉及golang的并发编程

服务器功能：从客户端接收数据，并且实现协程同步，将游戏结果输出在控制台
*/

//玩家信息结构体
type PlayerInfo struct {
	allScore   int    //玩家的总分
	sleepTime  uint16    //玩家线程在本轮的睡眠时间
	choice     uint16   //玩家的选择
	nameString string //玩家姓名
}

//线程同步通道
var ch = make(chan bool)  //裁判线程通知主线程结束的通道
var ch1 = make(chan bool) //裁判线程通知玩家线程可以继续的通道
var wg = sync.WaitGroup{} //裁判线程等待玩家线程

//玩家线程
func playerTcp(player *PlayerInfo) {
	//绑定Tcp连接，本机10086端口
	listener,err := net.Listen("tcp","localhost:10086")
	if err != nil{
		fmt.Printf("tcp listener error, os exit……")
		os.Exit(-1)
	}
	conn, err := listener.Accept()
	if err!=nil{
		fmt.Printf("accept error, os exit……")
		os.Exit(-1)
	}

	//开始循环获取数据
	for i := 0; i < 10; i++ {
		//等待被通知开始
		<-ch1

		//从客户端获取睡眠时间
		numOne := make([]byte,2)
		_,err := conn.Read(numOne)
		if err != nil{
			fmt.Printf("tcp sleepTime error, os exit……")
			os.Exit(-1)
		}
		player.sleepTime = binary.LittleEndian.Uint16(numOne)

		//从客户端获取选择
		numTwo := make([]byte,2)
		_,err = conn.Read(numTwo)
		if err != nil{
			fmt.Printf("tcp choice error, os exit……")
			os.Exit(-1)
		}
		player.choice = binary.LittleEndian.Uint16(numTwo)

		//当前轮次结束
		wg.Done()
	}
}

func playerUdp(player *PlayerInfo)  {
	//绑定Udp连接，本机10087端口
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10087")
	if err != nil{
		fmt.Printf("udp listener error, os exit……")
		os.Exit(-1)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("ListenUDP err:", err)
		return
	}

	//开始循环获取数据
	for i := 0; i < 10; i++ {
		//等待被通知开始
		<-ch1

		//从客户端获取睡眠时间
		bytesOne := make([]byte,2)
		_,err := conn.Read(bytesOne)
		if err != nil{
			fmt.Printf("udp recv error\n")
		}
		player.sleepTime = binary.LittleEndian.Uint16(bytesOne)

		//从客户端获取选择
		bytesTwo := make([]byte,2)
		_,err = conn.Read(bytesTwo)
		if err != nil{
			fmt.Printf("udp recv error\n")
		}
		player.choice = binary.LittleEndian.Uint16(bytesTwo)

		//当前轮次结束
		wg.Done()
	}
}

//裁判线程
func judge(player1, player2 *PlayerInfo) {
	//输出表头
	fmt.Printf("\t\t%s\t\t\t\t\t\t%s\n"+
		"round\tsleeptime\tchoice\t\tsocre\tsleeptime\tchoice\t\tscore\n",
		player1.nameString, player2.nameString)

	//创建代表石头剪刀布的字符串数组
	var arr  = [3]string{"rock    ","scissors","cloth   "}

	//开始游戏（循环判断输赢
	for i := 1; i <= 10; i++ {
		//线程同步
		wg.Add(2)
		ch1 <- true
		ch1 <- true
		wg.Wait()

		//判断输赢
		if (player1.choice + 1) % 3 == player2.choice{
			//玩家1获胜
			player1.allScore += 2
			fmt.Printf("%d\t\t%d\t\t\t%s\t%d\t\t%d\t\t\t%s\t%d\n",
				i,player1.sleepTime,arr[player1.choice],2,player2.sleepTime,arr[player2.choice],0)
		}else if(player2.choice + 1)%3 == player1.choice{
			//玩家2获胜
			player2.allScore += 2
			fmt.Printf("%d\t\t%d\t\t\t%s\t%d\t\t%d\t\t\t%s\t%d\n",
				i,player1.sleepTime,arr[player1.choice],0,player2.sleepTime,arr[player2.choice],2)
		}else {
			//平局
			player1.allScore += 1
			player2.allScore += 1
			fmt.Printf("%d\t\t%d\t\t\t%s\t%d\t\t%d\t\t\t%s\t%d\n",
				i, player1.sleepTime, arr[player1.choice], 1, player2.sleepTime, arr[player2.choice], 1)
		}
	}
	//输出最后的比赛信息
	fmt.Printf("玩家%s得分：%d\n", player1.nameString, player1.allScore)
	fmt.Printf("玩家%s得分：%d\n", player2.nameString, player2.allScore)
	if player1.allScore > player2.allScore {
		fmt.Printf("恭喜%s获胜！！！", player1.nameString)
	}else if player1.allScore < player2.allScore {
		fmt.Printf("恭喜%s获胜！！！", player2.nameString)
	} else{
		fmt.Printf("平局！！！")
	}

	//游戏结束，关闭通道1，向通道2发送结束信息
	close(ch1)
	ch <- true
}

func main() {
	//创建玩家信息
	var player1 = PlayerInfo{0, 0, 0, "player1"}
	var player2 = PlayerInfo{0, 0, 0, "player2"}

	//创建三个线程
	go judge(&player1, &player2)
	go playerTcp(&player1)
	go playerUdp(&player2)

	//等待线程(裁判线程)结束
	<-ch
}