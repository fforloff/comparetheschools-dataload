package models

import "gopkg.in/mgo.v2/bson"

type (

	//Result ...
	Result struct {
		Year                       int           `bson:"year"`
		NoVceStudiesUnit34         int           `bson:"no_vce_studies_unit_3_4" csv:"no_vce_studies_unit_3_4"`
		NoVetCertificates          int           `bson:"no_vet_certificates" csv:"no_vet_certificates"`
		NoVCEStudents              int           `bson:"no_vce_students" csv:"no_vce_students"`
		NoVETStudents              int           `bson:"no_vet_students" csv:"no_vet_students"`
		NoVCALStudents             int           `bson:"no_vcal_students" csv:"no_vcal_students"`
		PercentVCEStudentsApplyUni int           `bson:"percent_vce_students_apply_uni" csv:"percent_vce_students_apply_uni"`
		PercentCompletionVCE       int           `bson:"percent_completion_vce" csv:"percent_completion_vce"`
		NoVCEBaccalaureate         int           `bson:"no_vce_baccalaureate" csv:"no_vce_baccalaureate"`
		PercentVETUnitsCompleted   int           `bson:"percent_vet_units_completed" csv:"percent_vet_units_completed"`
		PercentVCALUnitsCompleted  int           `bson:"percent_vcal_units_completed" csv:"percent_vcal_units_completed"`
		MedianVCEScore             int           `bson:"median_vce_score" csv:"median_vce_score"`
		PercentScore40AndOver      float32       `bson:"percent_score_40_and_over" csv:"percent_score_40_and_over"`
		School                     bson.ObjectId `bson:"school_id" csv:"school_id"`
		RankingScore               float32       `bson:"ranking_score"`
		RankingScoreWMA            float32       `bson:"ranking_score_wma"`
		Rank                       int           `bson:"rank"`
	}

	//School ...
	School struct {
		Name      string        `bson:"name" csv:"name"`
		Adult     bool          `bson:"adult" csv:"adult"`
		Small     bool          `bson:"small" csv:"small"`
		InterBcl  bool          `bson:"inter_bcl" csv:"inter_bcl"`
		EduSector string        `bson:"edu_sector" csv:"edu_sector"`
		Type      string        `bson:"type" csv:"type"`
		Address   string        `bson:"address" csv:"address"`
		Locality  string        `bson:"locality" csv:"locality"`
		Postcode  string        `bson:"postcode" csv:"postcode"`
		State     string        `bson:"state" csv:"state"`
		ID        bson.ObjectId `bson:"_id,omitempty"`
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
