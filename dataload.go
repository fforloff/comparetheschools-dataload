package main

import (
	// Standard library packages
	"encoding/csv"
	"errors"
	"flag"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/vint_au/comparetheschools-dataload/models"
)

func check(m string, e error) {
	if e != nil {
		log.Println(m + ": " + e.Error())
		os.Exit(1)
	}
}

type headerSlice []string

func mapStructToCSVColumns(hs headerSlice, strct interface{}) map[string]int {
	m := make(map[string]int)
	t := reflect.TypeOf(strct).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("csv")
		fieldName := t.Field(i).Name
		colnumber, found := hs.columnNumByHeader(tag)
		if found {
			m[fieldName] = colnumber
		}
	}
	return m
}

func (hs headerSlice) columnNumByHeader(title string) (int, bool) {
	for i, v := range hs {
		if v == title {
			return i, true
		}
	}
	return -1, false
}

func populateStructFromBody(row []string, m map[string]int, strct interface{}, debug bool) (interface{}, error) {
	if debug {
		log.Println(row)
	}
	t := reflect.TypeOf(strct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Input param is not a struct")
	}
	rr := reflect.New(t).Elem()

	for k, v := range m {
		if debug {
			log.Println(k)
			log.Println(row[v])
		}
		re := regexp.MustCompile(`\^|-|I\/D|N\/A|<4|< 4`)
		row[v] = re.ReplaceAllString(row[v], "")
		if debug {
			log.Println("after conversion")
			log.Println(row[v])
		}
		f := rr.FieldByName(k)
		switch f.Type().String() {
		case "int":
			switch row[v] {
			case "":
				f.SetInt(0)
			default:
				intval, err := strconv.ParseInt(row[v], 10, 0)
				if err != nil {
					return nil, errors.New("Cant convert to Int")
				}
				f.SetInt(intval)
			}
		case "float32":
			switch row[v] {
			case "":
				f.SetFloat(0)
			default:
				floatval, err := strconv.ParseFloat(row[v], 32)
				if err != nil {
					return nil, errors.New("Can't convert to Float")
				}
				f.SetFloat(floatval)
			}
		case "bool":
			switch row[v] {
			case "A":
				f.SetBool(true)
			case "*":
				f.SetBool(true)
			case "Y":
				f.SetBool(true)
			default:
				f.SetBool(false)
			}
		case "string":
			row[v] = strings.Title(strings.ToLower(row[v]))
			f.SetString(row[v])
		default:
			return nil, errors.New("Something is wrong...")
		}
	}
	return rr.Interface(), nil
}

func readCSV(f string) ([][]string, error) {
	var err error
	file, err := os.Open(f)
	check("Can't open file", err)
	defer file.Close()
	reader := csv.NewReader(file)
	dat, err := reader.ReadAll()
	return dat, err
}

func main() {
	filePtr := flag.String("in", "", "csv file to load")
	yearPtr := flag.Int("year", 0, "year of the results")
	mongoPtr := flag.String("mongo", "mongodb://localhost", "mongodb connection string")
	databasePtr := flag.String("database", "", "database name")
	debugPtr := flag.Bool("debug", false, "[false]|true")
	wmaPeriod := 3
	flag.Parse()
	if *filePtr == "" || *yearPtr == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	var err error
	//file := "/Users/vint/code/comparetheschools/raw_data/tabula-postcompletiondata-2015schools.csv"
	dat, err := readCSV(*filePtr)
	check("Can not read csv file", err)
	header := headerSlice(dat[0])
	if *debugPtr {
		log.Println(header)
	}
	body := dat[1:]
	schoolStructFieldToColumnsMap := mapStructToCSVColumns(header, &models.School{})
	//log.Println(schoolStructFieldToColumnsMap)
	resultStructFieldToColumnsMap := mapStructToCSVColumns(header, &models.Result{})
	//log.Println(resultStructFieldToColumnsMap)

	s, err := mgo.Dial(*mongoPtr)
	check("Can't connect to the database", err)

	sc := s.DB(*databasePtr).C("schools")
	rc := s.DB(*databasePtr).C("results")
	count := len(body)
	bar := pb.StartNew(count)
	for _, row := range body {
		bar.Increment()
		if *debugPtr {
			log.Println(row)
		}
		interfaceSchool, _ := populateStructFromBody(row, schoolStructFieldToColumnsMap, &models.School{}, *debugPtr)
		interfaceResult, _ := populateStructFromBody(row, resultStructFieldToColumnsMap, &models.Result{}, *debugPtr)
		school := interfaceSchool.(models.School)
		result := interfaceResult.(models.Result)
		query := bson.M{"name": school.Name, "locality": school.Locality}
		//_, err = sc.Upsert(bson.M{"$and": []interface{}{bson.M{"name": school.Name}, bson.M{"locality": school.Locality}}}, bson.M{"$set": school})

		_, err = sc.Upsert(query, bson.M{"$set": school})
		//_, err = sc.Upsert(bson.M{"name": school.Name, "locality": school.Locality}, bson.M{"$set": school})
		check("Can not upsert school", err)

		var sch models.School
		err = sc.Find(query).One(&sch)
		//err = sc.Find(bson.M{"$and": []interface{}{bson.M{"name": school.Name}, bson.M{"locality": school.Locality}}}).One(&sch)
		check("Can not find a school", err)

		result.School = sch.Id
		result.Year = *yearPtr
		result.RankingScore = result.CalculateRankingScore()
		//log.Println(result.RankingScore)

		var Scores []float32
		var rr []models.Result
		err = rc.Find(bson.M{"school_id": result.School}).Sort("year").Limit(wmaPeriod - 1).All(&rr)
		check("Can not perform the find query", err)
		for _, r := range rr {
			//log.Printf("Year: %v", r.Year)
			Scores = append(Scores, r.RankingScore)
		}
		Scores = append(Scores, result.RankingScore)
		result.RankingScoreWMA = result.CalculateRankingScoreWMA(Scores)

		_, err = rc.Upsert(bson.M{"school_id": result.School, "year": result.Year}, bson.M{"$set": result})
		//_, err = rc.Upsert(bson.M{"and": []bson.M{bson.M{"school_id": result.School}, bson.M{"year": result.Year}}}, bson.M{"$set": result})
		check("Can not upsert the result", err)

		//log.Println(school)

	}
	bar.FinishPrint("")

	//Calculate school ranks
	var rank int
	rank = 0
	var previousRankingScore float32
	previousRankingScore = 0.00
	var rrr []models.Result
	err = rc.Find(bson.M{"year": *yearPtr}).Sort("-ranking_score_wma").All(&rrr)
	check("Can not get results from the database.", err)

	for i, r := range rrr {
		if r.RankingScoreWMA != previousRankingScore {
			rank = i + 1
		}
		log.Printf("School ID: %v, Ranking score: %v, Rank: %v", r.School, r.RankingScoreWMA, rank)
		err = rc.Update(bson.M{"school_id": r.School, "year": r.Year}, bson.M{"$set": bson.M{"rank": rank}})
		//check("Can not update rank.", err)
		previousRankingScore = r.RankingScoreWMA
	}

}
