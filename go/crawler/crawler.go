package main

import (
	"fmt"
	"net/http"
	"time"
	"sort"
	"regexp"
	"github.com/PuerkitoBio/goquery"

	"database/sql"
	_"github.com/mattn/go-sqlite3"
)

func getIntListDiff(listA []int, listB []int) ([]int) {
	largeList := listA
	smallList := listB
	if len(largeList) < len(smallList) {
		tmp := largeList
		largeList = smallList
		smallList = tmp
	}

	// ソートして調べる
	sort.Slice(largeList,func(i, j int) bool { return largeList[i] < largeList[j] })
	sort.Slice(smallList,func(i, j int) bool { return smallList[i] < smallList[j] })
	isSameList := make([]bool, len(largeList)) // init all false
	for i,_ := range largeList {
		for j := 0; j < len(smallList); j++ {
			if smallList[j] == largeList[i] {
				isSameList[i] = true
				break
			}
		}
	}
	var diffList []int
	for i,isSame := range isSameList {
		if !isSame {
			diffList = append(diffList, largeList[i])
		}
	}
	return diffList
}

// <---- modeling ---->
const (
	db_path = "./"
	db_name = "test.db"
)

type db_data struct {
	id int
	word string
	articleIDs []int // articleID,articleID,articleID,...
}

func DBinit() (error){
	// 正常に生成出来たらnilを返す
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`CREATE TABLE "words" ("id" integer primary key not null, "word" text)`,
	)
	_, err = db.Exec(
		`CREATE TABLE "article_words" ("word_id" integer not null, "article_id" integer not null, primary key(article_id,word_id))`,
	)
	if err != nil {
		if _,err = db.Exec(`delete from words`);err!=nil {
			return err
		}
		if _,err = db.Exec(`delete from article_words`);err!=nil {
			return err
		}
	}
	return nil	
}

func getAllWordInArticle() (string, error) {
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return "", err
	}

	res, err := db.Query(
		`select * from article_words`,
	)
	if err != nil {
		return "", err
	}
	for res.Next() {
		var word_id int
		var article_id int
		if err := res.Scan(&word_id, &article_id);err!=nil {
			return "",err
		}
		fmt.Printf("word_id:%d, article_id:%d\n", word_id,article_id)
	}
	return "", err
}

func getArticleIDs(data *db_data) (error) {
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}
	res, err := db.Query(
		`select article_id from article_words where word_id=?`,
		data.id,
	)
	if err != nil {
		return err
	}
	for res.Next() {
		var article_id int
		if err := res.Scan(&article_id);err!=nil {
			return err
		}
		data.articleIDs = append(data.articleIDs, article_id)
	}
	return nil
}

func insertNewWord(data *db_data) (error){
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}

	res, err := db.Exec(
		`insert into words (word) values (?)`,
		data.word,
	)
	if err != nil {
		return err
	}
	word_id,err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`insert into article_words (word_id, article_id) values (?,?)`,
		word_id,
		data.articleIDs[0],
	)
	return nil
}

func updateArticleID(data *db_data) (error){
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}

	res := db.QueryRow(
		`SELECT * FROM words WHERE word=?`,
		data.word,
	)

	err = res.Scan(&data.id)
	if err!=nil {
		return err
	}
	storeStruct := data

	err = getArticleIDs(storeStruct)
	if err!=nil {
		return err
	}


	_, err = db.Exec(
		`insert article_words (word_id,article_id) value (?,?)`,
		data.articleIDs[len(data.articleIDs)-1],
		data.id,
	)
	if err != nil {
		return err
	}
	return nil
}

//<---- crawler ---->

func get_a_article(url string)string {
	resp, err := http.Get(url)
	if err != nil{
		panic(err)
	}
	defer resp.Body.Close()

	/*
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	*/

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil{
		panic(err)
	}

	retrunText := ""

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		rawText := s.Text()
		if !regexp.MustCompile("[0-9]+年").MatchString(rawText) {
			re := regexp.MustCompile("[[0-9]+]")
			context := re.ReplaceAllString(rawText, "")
			re = regexp.MustCompile("\n")
			context = re.ReplaceAllString(context, "")
			retrunText += context
		}
	})
	return retrunText
}

func scrape() {
	urlList := [2]string{
		"https://ja.wikipedia.org/wiki/%E6%9C%88%E3%83%8E%E7%BE%8E%E5%85%8E", //月ノ美兎
		"https://ja.wikipedia.org/wiki/%E5%90%8D%E5%8F%96%E3%81%95%E3%81%AA", //名取さな
	}

	for i := range urlList {
		fmt.Printf("%s : %s\n", urlList[i], get_a_article(urlList[i]))
	}
}

func main() {
	if err := DBinit(); err!=nil {
		panic(err)
	}
	
	a_d := db_data{word:"hoge"}
	a_d.articleIDs = append(a_d.articleIDs, 0)
	if err := insertNewWord(&a_d);err!=nil {
		panic(err)
	}
	a_d.articleIDs = append(a_d.articleIDs, 1)
	if err := updateArticleID(&a_d);err!=nil {
		panic(err)
	}

	if _,err := getAllWordInArticle();err!=nil {
		panic(err)
	}

	fmt.Printf("get ready!")
	for {
		scrape()
		fmt.Printf("the end!")
		time.Sleep(1*time.Hour)
	}
}