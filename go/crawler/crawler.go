package main

import (
	"fmt"
	"net/http"
	"time"
	"sort"
	"regexp"
	"strings"
	"github.com/PuerkitoBio/goquery"

	"database/sql"
	"github.com/jinzhu/copier"
	"github.com/bluele/mecab-golang"
	_"github.com/mattn/go-sqlite3"
)

func getIntListDiff(listA []int, listB []int) ([]int) {
	// listAに存在してlistBに存在しない要素の取得関数
	largeList := listA
	smallList := listB

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
	defer db.Close()

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
	defer db.Close()

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

func getWordID(data *db_data) (error) {
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}
	defer db.Close()

	res := db.QueryRow(
		`SELECT * FROM words WHERE word=?`,
		data.word,
	)
	err = res.Scan(&data.id)
	if err!=nil {
		return err
	}
	return nil
}

func getArticleIDs(data *db_data) (error) {
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Query(
		`select article_id from article_words where word_id=?`,
		data.id,
	)
	if err != nil {
		return err
	}
	data.articleIDs = nil
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
	defer db.Close()

	// 単語を追加する
	res, err := db.Exec(
		`insert into words (word) values (?)`,
		data.word,
	)
	if err != nil {
		return err
	}
	// 追加したidを取得して記事と結びつける
	word_id,err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`insert into article_words (word_id, article_id) values (?,?)`,
		word_id,
		data.articleIDs[0],
	)
	if err!= nil {
		return err
	}
	return nil
}

func updateArticleID(data *db_data) (error){
	db, err := sql.Open("sqlite3", db_path+db_name)
	if err != nil {
		return err
	}
	defer db.Close()

	res := db.QueryRow(
		`SELECT id FROM words WHERE word=?`,
		data.word,
	)

	err = res.Scan(&data.id)
	if err!=nil {
		return err
	}
	storeStruct := db_data{}
	copier.Copy(&storeStruct, data) // deep copy
	
	err = getArticleIDs(&storeStruct)
	if err!=nil {
		return err
	}

	diff := getIntListDiff(data.articleIDs, storeStruct.articleIDs)
	for _,articleId := range diff {
		_, err = db.Exec(
			`insert into article_words (word_id,article_id) values (?,?)`,
			data.id,
			articleId,
		)
		if err != nil {
			return err
		}
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
		
		re := regexp.MustCompile("[[0-9]+]")
		context := re.ReplaceAllString(rawText, "")
		re = regexp.MustCompile("[0-9]+年")
		context = re.ReplaceAllString(context, "")
		re = regexp.MustCompile("\n")
		context = re.ReplaceAllString(context, "")
		retrunText += context
	})
	return retrunText
}

func scrape() string {
	urlList := [2]string{
		"https://ja.wikipedia.org/wiki/%E6%9C%88%E3%83%8E%E7%BE%8E%E5%85%8E", //月ノ美兎
		"https://ja.wikipedia.org/wiki/%E5%90%8D%E5%8F%96%E3%81%95%E3%81%AA", //名取さな
	}

	for i := range urlList {
		return get_a_article(urlList[i])
	}
	return ""
}

//<---- word controller ---->
func splitArticleToSentence(article string) []string {
	return strings.Split(article,"。")
}

func splitSentenceToArticle(sentence string) ([]string, error) {
	m,err := mecab.New("-Owakati")
	if err != nil {
		return nil,err
	}
	defer m.Destroy()

	tg, err := m.NewTagger()
	if err != nil {
		return nil,err
	}
	defer tag.Destroy()

	lt, err := m.NewLattice(sentence)
	if err != nil {
		return nil,err
	}
	defer lt.Destroy()

	wordList := []string
	node := tg.ParseToNode(lt)
	for {
		features := strings.Split(node.Feature(),",")
		if features[0] != "BOS/EOS" {
			fmt.Printf("%s %s\n",node.Surface(), node.Feature())
		}
		if node.Next() != nil {
			break
		}
	}
	return wordList
}

func main() {
	if err := DBinit(); err!=nil {
		panic(err)
	}

	fmt.Printf("get ready!\n")
	for {
		for _,list := range splitSentenceToArticle(splitArticleToSentence(scrape())[0]) {
			fmt.Println(list)
		}
		fmt.Printf("the end!\n")
		time.Sleep(1*time.Hour)
	}
}