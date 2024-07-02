package db

// Example of a phishing record:
//
//	{
//		"phish_id": 123456,
//		"url": "https://www.example.com/",
//		"phish_detail_url": "http://www.phishtank.com/phish_detail.php?phish_id=123456",
//		"submission_time": "2009-06-19T15:15:47+00:00",
//		"verified": "yes",
//		"verification_time": "2009-06-19T15:37:31+00:00",
//		"online": "yes",
//		"target": "1st National Example Bank",
//		"details": {
//			[
//				"ip_address": "1.2.3.4",
//				"cidr_block": "1.2.3.0/24",
//				"announcing_network": "1234",
//				"rir": "arin",
//				"detail_time": "2006-10-01T02:30:54+00:00"
//			]
//		}
//	}
type PhishingRecord struct {
	PhishID            int    `json:"phish_id" bson:"phish_id"`
	Url                string `json:"url" bson:"url"`
	PhishDetailUrl     string `json:"phish_detail_url" bson:"phish_detail_url"`
	SubmissionTime     string `json:"submission_time" bson:"submission_time"`
	SubmissionTimeUnix int64  `bson:"submission_time_unix"`
	Verified           string `json:"verified" bson:"verified"`
	VerificationTime   string `json:"verification_time" bson:"verification_time"`
	Online             string `json:"online" bson:"online"`
	Target             string `json:"target" bson:"target"`
	Details            []struct {
		IpAddress         string `json:"ip_address" bson:"ip_address"`
		CidrBlock         string `json:"cidr_block" bson:"cidr_block"`
		AnnouncingNetwork string `json:"announcing_network" bson:"announcing_network"`
		Rir               string `json:"rir" bson:"rir"`
		DetailTime        string `json:"detail_time" bson:"detail_time"`
	} `json:"details" bson:"details"`
}
