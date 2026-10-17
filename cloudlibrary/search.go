package cloudlibrary

import (
	"net/http"
	"net/url"
)

type Category struct {
	CategoryID        string `json:"categoryId"`        // categoryId: "a389w"
	DirectMemberCount int    `json:"directMemberCount"` // directMemberCount: 3647
	I18nKey           string `json:"i18nKey"`           // i18nKey: "category.BisacSubjectHeading.FIC056000"
	ID                int    `json:"id"`                // id: 7552
	Name              string `json:"name"`              // name: "Hispanic & Latino"
	ParentID          string `json:"parentId"`          // parentId: "4q9w"
	SubCategoryCount  int    `json:"subCategoryCount"`  // subCategoryCount: 0
}

// https://ebook.yourcloudlibrary.com/library/guelphpl/search?title=legends+and+lattes&authors=&series=&isbn=&category=&format=&available=any&language=&sort=&orderBy=relevence&owned=yes&_data=routes%2Flibrary.%24name.search

type SearchResults struct {
	Items         []SearchItem `json:"items"`
	Query         string       `json:"query"`
	TotalItems    int          `json:"totalItems"`
	TotalSegments int          `json:"totalSegments"`
}

type SearchItem struct {
	AgeClassification       int      `json:"ageClassification"`       // ageClassification: 1
	Authors                 []string `json:"authors"`                 // authors: ["Donlay, Philip S.", "Kennedy, Chris"]
	BibliographicIdentifier string   `json:"bibliographicIdentifier"` // bibliographicIdentifier: "g7a16bop8"
	CanBorrow               bool     `json:"canBorrow"`               // canBorrow: true
	CanHold                 bool     `json:"canHold"`                 // canHold: false
	CanListen               bool     `json:"canListen"`               // canListen: false
	CanRead                 bool     `json:"canRead"`                 // canRead: false
	CanRemoveHold           bool     `json:"canRemoveHold"`           // canRemoveHold: false
	CanRenew                bool     `json:"canRenew"`                // canRenew: false
	CanReturn               bool     `json:"canReturn"`               // canReturn: false
	CanSave                 bool     `json:"canSave"`                 // canSave: true
	CanSuggest              bool     `json:"canSuggest"`              // canSuggest: false
	CatalogItemId           string   `json:"catalogItemId"`           // catalogItemId: "dig-bibg-clg7a16bop8"
	// Contributors            any      `json:"contributors"`            // contributors: [{name: "Donlay, Philip S.", role: 0}, {name: "Kennedy, Chris", role: 1}]
	Crumbtrail             string   `json:"crumbtrail"`             // crumbtrail: "id=g7a16bop8&owned=yes&available=any"
	CurrentlyAvailable     int      `json:"currentlyAvailable"`     // currentlyAvailable: 1
	CurrentlyLoaned        int      `json:"currentlyLoaned"`        // currentlyLoaned: 0
	CurrentlyReserved      int      `json:"currentlyReserved"`      // currentlyReserved: 0
	DateAdded              string   `json:"dateAdded"`              // dateAdded: "2014-05-02"
	DatePublished          string   `json:"datePublished"`          // datePublished: "2005-05-15"
	DocumentId             string   `json:"documentId"`             // documentId: "ezscfz9"
	Duration               int      `json:"duration"`               // duration: 34125000
	EpubFormat             int      `json:"epubFormat"`             // epubFormat: 0
	Format                 string   `json:"format"`                 // format: "audio"
	Id                     string   `json:"id"`                     // id: "g7a16bop8"
	ImageLinkThumbnail     string   `json:"imageLinkThumbnail"`     // imageLinkThumbnail: "https://images.yourcloudlibrary.com/delivery/img?type=DOCUMENTIMAGE&documentID=ezscfz9&size=MEDIUM"
	ImprintName            string   `json:"imprintName"`            // imprintName: ""
	IsHeld                 bool     `json:"isHeld"`                 // isHeld: false
	IsHoldAvailable        bool     `json:"isHoldAvailable"`        // isHoldAvailable: false
	IsPayPerUse            bool     `json:"isPayPerUse"`            // isPayPerUse: false
	IsPreSale              bool     `json:"isPreSale"`              // isPreSale: false
	IsSuggested            bool     `json:"isSuggested"`            // isSuggested: false
	Isbn                   string   `json:"isbn"`                   // isbn: "9781596073906"
	Language               string   `json:"language"`               // language: "eng"
	MatchingScore          float32  `json:"matchingScore"`          // matchingScore: 5.544257
	PpuTitleExcludes       []any    `json:"ppuTitleExcludes"`       // ppuTitleExcludes: []
	ProductFormDescription any      `json:"productFormDescription"` // productFormDescription: null
	PublisherName          string   `json:"publisherName"`          // publisherName: "Books in Motion"
	Rating                 int      `json:"rating"`                 // rating: 0
	SeriesTitle            string   `json:"seriesTitle"`            // seriesTitle: ""
	Size                   int      `json:"size"`                   // size: 273008919
	Subjects               []string `json:"subjects"`               // subjects: ["FIC031000", "FIC000000"]
	Subtitle               string   `json:"subtitle"`               // subtitle: ""
	Summary                string   `json:"summary"`                // summary: "In the Atlantic Ocean, hurricane Helena is gathering strength, becoming the most powerful storm in recorded history. As Helena bears down on Bermuda, Donovan Nash, along with other members of the scientific research organization Eco-Watch, are called to fly in and extract key government people who h..."
	Title                  string   `json:"title"`                  // title: "Category Five"
	TotalCopies            int      `json:"totalCopies"`            // totalCopies: 1
	TotalExtents           any      `json:"totalExtents"`           // totalExtents: null
	YearPublished          int      `json:"yearPublished"`          // yearPublished: 2005
}

func (c *Client) Search(query string) (*SearchResults, error) {
	q := url.Values{}
	q.Add("query", query)
	q.Add("format", "")
	q.Add("available", "any")
	q.Add("language", "")
	q.Add("sort", "")
	q.Add("orderBy", "relevence")
	q.Add("owned", "yes")
	q.Add("_data", "routes/library.$name.search")

	r := &SearchResponse{}
	err := c.request(http.MethodGet, "https://ebook.yourcloudlibrary.com/library/guelphpl/search?"+q.Encode(), nil, 1, r)
	if err != nil {
		return nil, err
	}
	return &r.Results.Search, nil
}
