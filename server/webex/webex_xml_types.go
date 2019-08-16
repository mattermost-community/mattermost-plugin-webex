package webex

import "encoding/xml"

// PMRR = PersonalMeetingRoomResponse
type GetPMRR struct {
	XMLName xml.Name    `xml:"message"`
	Header  Header      `xml:"header"`
	Body    GetPMRRBody `xml:"body"`
}

type GetPMRRBody struct {
	BodyContent GetPMRRBodyContent `xml:"bodyContent"`
}

type GetPMRRBodyContent struct {
	XMLName             xml.Name      `xml:"bodyContent"`
	Avatar              GetPMRRAvatar `xml:"avatar"`
	PersonalMeetingRoom PMR           `xml:"personalMeetingRoom"`
	//HostMeetingURL string        `xml:"hostMeetingURL"`
}

type GetPMRRAvatar struct {
	//XMLName          xml.Name `xml:"avatar"`
	Url              string `xml:"url"`
	LastModifiedTime int    `xml:"lastModifiedTime"`
	IsUploaded       bool   `xml:"isUploaded"`
}

type PMR struct {
	XMLName    xml.Name `xml:"personalMeetingRoom"`
	Title      string   `xml:"title"`
	PMRUrl     string   `xml:"personalMeetingRoomURL"`
	AccessCode string   `xml:"accessCode"`
}

type Header struct {
	XMLName  xml.Name `xml:"header"`
	Response Response `xml:"response"`
}

type Response struct {
	XMLName xml.Name `xml:"response"`
	Result  string   `xml:"result"`
}
