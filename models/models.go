package models

import "gopkg.in/mgo.v2/bson"

type (
	//Output ...
	Output struct {
		Result `bson:",inline"`
		School `bson:",inline"`
	}

	//Result ...
	Result struct {
		Year                       int           `json:"year" bson:"year"`
		NoVceStudiesUnit34         int           `json:"no_vce_studies_unit_3_4" bson:"no_vce_studies_unit_3_4" csv:"no_vce_studies_unit_3_4"`
		NoVetCertificates          int           `json:"no_vet_certificates" bson:"no_vet_certificates" csv:"no_vet_certificates"`
		NoVCEStudents              int           `json:"no_vce_students" bson:"no_vce_students" csv:"no_vce_students"`
		NoVETStudents              int           `json:"no_vet_students" bson:"no_vet_students" csv:"no_vet_students"`
		NoVCALStudents             int           `json:"no_vcal_students" bson:"no_vcal_students" csv:"no_vcal_students"`
		PercentVCEStudentsApplyUni int           `json:"percent_vce_students_apply_uni" bson:"percent_vce_students_apply_uni" csv:"percent_vce_students_apply_uni"`
		PercentCompletionVCE       int           `json:"percent_completion_vce" bson:"percent_completion_vce" csv:"percent_completion_vce"`
		NoVCEBaccalaureate         int           `json:"no_vce_baccalaureate" bson:"no_vce_baccalaureate" csv:"no_vce_baccalaureate"`
		PercentVETUnitsCompleted   int           `json:"percent_vet_units_completed" bson:"percent_vet_units_completed" csv:"percent_vet_units_completed"`
		PercentVCALUnitsCompleted  int           `json:"percent_vcal_units_completed" bson:"percent_vcal_units_completed" csv:"percent_vcal_units_completed"`
		MedianVCEScore             int           `json:"median_vce_score" bson:"median_vce_score" csv:"median_vce_score"`
		PercentScore40AndOver      float32       `json:"percent_score_40_and_over" bson:"percent_score_40_and_over" csv:"percent_score_40_and_over"`
		School                     bson.ObjectId `json:"school" bson:"school_id" csv:"school_id"`
		RankingScore               float32       `json:"ranking_score" bson:"ranking_score"`
		RankingScoreWMA            float32       `json:"ranking_score_wma" bson:"ranking_score_wma"`
		Rank                       int           `json:"rank" bson:"rank"`
	}

	//School ...
	School struct {
		Name      string        `json:"school_name" bson:"name" csv:"name"`
		Adult     bool          `json:"adult" csv:"adult"`
		Small     bool          `json:"small" csv:"small"`
		InterBcl  bool          `json:"inter_bcl" csv:"inter_bcl"`
		EduSector string        `json:"edu_sector" csv:"edu_sector"`
		Type      string        `json:"type" csv:"type"`
		Address   string        `json:"address" csv:"address"`
		Locality  string        `json:"locality" csv:"locality"`
		Postcode  string        `json:"postcode" csv:"postcode"`
		State     string        `json:"state" csv:"state"`
		Id        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	}
)

//CalculateRankingScore is a method to calculate absolute ranking score
// for a given school at a given year
func (r Result) CalculateRankingScore() (score float32) {
	score = float32((r.MedianVCEScore+1)*(r.PercentCompletionVCE+1)) * (r.PercentScore40AndOver + float32(1))
	return
}

//CalculateRankingScoreWMA is a method to calculate a weighted moving average of the
//ranking score over number of years
func (r Result) CalculateRankingScoreWMA(ss []float32) (result float32) {
	var numerator float32
	var denominator int
	for i, v := range ss {
		numerator = numerator + float32(i+1)*v
		denominator = denominator + i + 1
	}
	result = numerator / float32(denominator)
	return
}
