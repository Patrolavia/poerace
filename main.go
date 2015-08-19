package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ladder struct {
	Rank      int
	Character string
	Class     string
	Level     int
	Exp       int64
	Dead      bool
}

func base(w http.ResponseWriter, r *http.Request, msg string) {
	fmt.Fprintf(w, `<html><head><meta charset="utf8"/></head><body>%s</body></html>`, msg)
}

func bad(w http.ResponseWriter, r *http.Request, msg string) {
	base(w, r, "發生錯誤："+msg)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head><meta charset="utf8"/><link rel="icon" href="http://ronmi.tw/favicon.ico" type="image/x-icon" /></head><body><img src="http://ronmi.tw/logo32.png" />This program is wrote by Patrolavia.<p><form action="p" method="POST">請到官網賽事公告頁面，在「<span style="color:red">匯出成 CSV 檔</span>」右鍵複製鏈結位址，然後在這裡貼上 <input name="url" /><br />請輸入你的角色名稱 <input name="id" /><br /><input type="submit" value="確定" /></form></body></html>`)
}

func do(w http.ResponseWriter, r *http.Request) {
	e := func(msg string) {
		bad(w, r, msg)
	}
	clsMap := map[string]string{
		"Duelist":  "決鬥者",
		"Shadow":   "暗影刺客",
		"Witch":    "女巫",
		"Templar":  "聖堂武士",
		"Marauder": "野蠻人",
		"Scion":    "貴族",
		"Ranger":   "遊俠",
	}
	u := r.PostFormValue("url")
	id := r.PostFormValue("id")

	if u == "" {
		e("請輸入網址")
		return
	}

	if id == "" {
		e("請輸入角色名稱")
		return
	}

	resp, err := http.Get(u)
	if err != nil {
		e("無法取得排名資料，請確認網址是否正確、網路連線是否正常，然後再試一次")
		return
	}

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e("成功連線到伺服器，但無法取得排名資料，請確認網址是否正確、網路連線是否正常，然後再試一次")
		return
	}

	strArr := strings.Split(string(rawData), "\n")
	// skip title line
	strArr = strArr[1:]

	data := make(map[string][]ladder)

	for k := range clsMap {
		data[k] = make([]ladder, 0)
	}

	// 總排名
	var userTotal ladder
	// 職業排名
	var userClass int

	for no := 0; no < len(strArr); no++ {
		if strArr[no] == "" {
			continue
		}
		aChar := strings.Split(strArr[no], ",")
		l := ladder{}
		l.Rank, _ = strconv.Atoi(aChar[1])
		l.Character = aChar[3]
		l.Class = aChar[4]
		l.Level, _ = strconv.Atoi(aChar[5])
		l.Exp, _ = strconv.ParseInt(aChar[6], 10, 64)
		l.Dead = aChar[7] != ""

		data[l.Class] = append(data[l.Class], l)
		if l.Character == id && userTotal.Rank == 0 {
			userTotal = l
			userClass = len(data[l.Class])
		}
	}

	if userTotal.Rank == 0 {
		e(fmt.Sprintf(`找不到 %s 的資料，請確認是否輸入正確，或是貼錯到別場 race 的網址`, id))
		return
	}

	ret := fmt.Sprintf(`
<h1>%s</h1>
<p>總排名: %d<br />
%s排名: %d</p>
<h1>各職業前 20 名</h1>
`, id, userTotal.Rank, clsMap[userTotal.Class], userClass)

	for cls, ladders := range data {
		ret += `<div style="float:left;margin:20px;border:1px solid black;padding:5px"><h3>` + clsMap[cls] + `</h3><table border="0" cellspacing="0" cellpadding="5">` + "\n"
		for k := 0; k < 20; k++ {
			if len(ladders) <= k {
				break
			}
			l := ladders[k]
			ret += fmt.Sprintf(`<tr><td style="border-right:1px solid black">%d</td><td style="border-right:1px solid black">%s</td><td style="border-right:1px solid black">總排名 %d</td><td style="border-right:1px solid black">等級 %d</td><td style="text-align:right">%d</td></tr>` + "\n",
				k+1, l.Character, l.Rank, l.Level, l.Exp)
		}
		ret += `</table></div>` + "\n"
	}

	base(w, r, ret)
}

func main() {
	http.HandleFunc("/p", do)
	http.HandleFunc("/", index)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		log.Fatalf("無法啟動: %s", err)
	}
}
