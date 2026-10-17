package cloudlibrary

import (
	"net/http"
	"net/url"
)

// {"patronId":"a3dac748-4624-420f-8928-c9faa8d82a47","patronItems":[{"id":"cl-7jcxcyo4mgYxsscCD","instanceId":"cl-7jcxcyo4mgYxsscCD","title":"Sixth of the Dusk","contributors":[{"name":"Sanderson, Brandon","role":"Other"}],"borrowedDate":"2026-10-17","dueDate":"2026-11-07 13:05:27","returnedDate":null,"holdAvailableDate":null,"holdExpiresDate":null,"priorityHoldDate":null,"productFormDescription":null,"holdQueuePosition":0,"isHoldAvailable":false,"isbn":"9781938570070","imageLink":"https://images.yourcloudlibrary.com/delivery/img?type=DOCUMENTIMAGE&documentID=ofo6v89&size=MEDIUM","itemId":"ofo6v89","loanId":"a9hqta89","bibliographicId":"7jcxcyo4m","catalogItemId":"dig-bibg-cl7jcxcyo4m","mediaCategory":"eBook","mediaType":"Epub","state":"Borrowed","isPreSale":false,"publishedDate":"2014-10-20","canFavorite":false,"canReturn":true,"canMarkRead":false,"canBorrow":false,"canRenew":false,"canHold":false,"canRemoveHold":false,"canSave":true,"canSuggest":false,"canRead":true,"canListen":true,"canBag":false,"canHide":false,"isFavorite":false,"isRead":false,"isHeld":false,"isBagged":false,"isSuggested":false,"isHidden":false}],"totalSegments":1,"currentSegment":1,"itemsPerSegment":20,"totalItems":1,"RPC_DOMAIN_PUBLIC":"https://service.yourcloudlibrary.com","reaktor":"kiTvTolG43k-CPHKNvW*-LgYEj*KNlyYXodDQajBrL-@"}
type MyBooksRespose struct {
	PatronItems     []PatronItem `json:"patronItems"`       //`:[]
	TotalSegments   int          `json:"totalSegments"`     //`:1
	CurrentSegment  int          `json:"currentSegment"`    //`:1
	ItemsPerSegment int          `json:"itemsPerSegment"`   //`:20
	TotalItems      int          `json:"totalItems"`        //`:0
	RpcDomainPublic string       `json:"RPC_DOMAIN_PUBLIC"` //`:"https://service.yourcloudlibrary.com"
	Reaktor         string       `json:"reaktor"`           //`:"kiTvTolG43k-CPHKNvW*-LgYEj*KNlyYXodDQajBrL-@"
}

type PatronItem struct {
	ID                     string                  `json:"id"`
	InstanceID             string                  `json:"instanceId"`
	Title                  string                  `json:"title"`
	Contributors           []PatronItemContributor `json:"contributors"`
	BorrowedDate           string                  `json:"borrowedDate"`
	DueDate                string                  `json:"dueDate"`
	ReturnedDate           *string                 `json:"returnedDate"`
	HoldAvailableDate      *string                 `json:"holdAvailableDate"`
	HoldExpiresDate        *string                 `json:"holdExpiresDate"`
	PriorityHoldDate       *string                 `json:"priorityHoldDate"`
	ProductFormDescription *string                 `json:"productFormDescription"`
	HoldQueuePosition      int                     `json:"holdQueuePosition"`
	IsHoldAvailable        bool                    `json:"isHoldAvailable"`
	ISBN                   string                  `json:"isbn"`
	ImageLink              string                  `json:"imageLink"`
	ItemID                 string                  `json:"itemId"`
	LoanID                 string                  `json:"loanId"`
	BibliographicID        string                  `json:"bibliographicId"`
	CatalogItemID          string                  `json:"catalogItemId"`
	MediaCategory          string                  `json:"mediaCategory"`
	MediaType              string                  `json:"mediaType"`
	State                  string                  `json:"state"`
	IsPreSale              bool                    `json:"isPreSale"`
	PublishedDate          string                  `json:"publishedDate"`
	CanFavorite            bool                    `json:"canFavorite"`
	CanReturn              bool                    `json:"canReturn"`
	CanMarkRead            bool                    `json:"canMarkRead"`
	CanBorrow              bool                    `json:"canBorrow"`
	CanRenew               bool                    `json:"canRenew"`
	CanHold                bool                    `json:"canHold"`
	CanRemoveHold          bool                    `json:"canRemoveHold"`
	CanSave                bool                    `json:"canSave"`
	CanSuggest             bool                    `json:"canSuggest"`
	CanRead                bool                    `json:"canRead"`
	CanListen              bool                    `json:"canListen"`
	CanBag                 bool                    `json:"canBag"`
	CanHide                bool                    `json:"canHide"`
	IsFavorite             bool                    `json:"isFavorite"`
	IsRead                 bool                    `json:"isRead"`
	IsHeld                 bool                    `json:"isHeld"`
	IsBagged               bool                    `json:"isBagged"`
	IsSuggested            bool                    `json:"isSuggested"`
	IsHidden               bool                    `json:"isHidden"`
}

type PatronItemContributor struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func (c *Client) MyBooks() ([]PatronItem, error) {
	q := url.Values{}
	q.Add("segment", "1")
	q.Add("pageSize", "20")
	q.Add("_data", "routes/library.$name.mybooks.current")

	body := url.Values{}
	body.Add("format", "")
	body.Add("sort", "BorrowedDateDescending")

	r := &MyBooksRespose{}
	err := c.request(http.MethodPost, "https://ebook.yourcloudlibrary.com/library/guelphpl/mybooks/current?"+q.Encode(), body, 1, r)
	if err != nil {
		return nil, err
	}

	c.rpcDomainPublic = r.RpcDomainPublic
	c.reaktor = r.Reaktor
	return r.PatronItems, nil
}
