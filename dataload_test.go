package main

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/fforloff/comparetheschools-dataload/models"
)

func TestCheckFunc(t *testing.T) {
	message := "Message to all"
	e := errors.New("I can't work in such poor conditions!")
	if os.Getenv("DO_CHECK") == "1" {
		check(message, e)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckFunc")
	cmd.Env = append(os.Environ(), "DO_CHECK=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

var hs = headerSlice{"name", "small", "locality", "adult", "inter_bcl", "median_vce_score", "percent_score_40_and_over", "percent_completion_vce"}
var headerMap = map[string]int{"Name": 0, "Small": 1, "Locality": 2, "Adult": 3, "InterBcl": 4}
var row = []string{"Springfield Elementary School", "Y", "Springfield", "A", "*", "10", "1.01", "16"}

func TestColumnNumberByHeader(t *testing.T) {
	colnumber, found := hs.columnNumByHeader("small")
	if colnumber != 1 && !found {
		t.Errorf("Expected colnumber 1, got %d; expected found to be 'true', got %t.", colnumber, found)
	}
	colnumber, found = hs.columnNumByHeader("big")
	if colnumber != -1 && found {
		t.Errorf("Expected colnumber -1, got %d; expected found to be 'false', got %t.", colnumber, found)
	}
}

func TestMapStructToCSVColumn(t *testing.T) {
	//	sm := &models.School{}
	//	log.Println(sm)
	schoolStructFieldToColumnsMap := mapStructToCSVColumns(hs, &models.School{})
	mapsEqual := reflect.DeepEqual(schoolStructFieldToColumnsMap, headerMap)
	if !mapsEqual {
		t.Errorf("Struct to Header Columns mismatch, expected %v, got %v", headerMap, schoolStructFieldToColumnsMap)
	}
}

func TestPopulateStructFromBody(t *testing.T) {
	interfaceSchool, _ := populateStructFromBody(row, mapStructToCSVColumns(hs, &models.School{}), &models.School{}, false)
	school := interfaceSchool.(models.School)
	if school.Name != "Springfield Elementary School" {
		t.Errorf("Unexpected result in the school name while populating struct from a row. Got %+v", school)
	}
	if school.Locality != "Springfield" {
		t.Errorf("Unexpected result in the locality while populating struct from a row. Got %+v", school)
	}
	if school.Small != true || school.Adult != true || school.InterBcl != true {
		t.Errorf("Unexpected result in the size/adult/international bacalaureate statuses while populating struct from a row. Got %+v", school)
	}
}

func TestCalculateRankingScore(t *testing.T) {
	// expecting (10+1)*(1.01+1)*(16+1) = 375.87
	expectedScore := float32(375.87)
	interfaceResult, _ := populateStructFromBody(row, mapStructToCSVColumns(hs, &models.Result{}), &models.Result{}, false)
	result := interfaceResult.(models.Result)
	score := result.CalculateRankingScore()
	if score != expectedScore {
		t.Errorf("Wrong score. Expected %v, got %v", expectedScore, score)
	}
}

func TestCalculateRankingScoreWMA(t *testing.T) {
	scores := []float32{10.0, 20.0, 30.0}
	expectedWMA := float32(23.333334)
	interfaceResult, _ := populateStructFromBody(row, mapStructToCSVColumns(hs, &models.Result{}), &models.Result{}, false)
	result := interfaceResult.(models.Result)
	wma := result.CalculateRankingScoreWMA(scores)
	if wma != expectedWMA {
		t.Errorf("Wrong Ranking Score WMA. Expected %v, got %v", expectedWMA, wma)
	}

}
