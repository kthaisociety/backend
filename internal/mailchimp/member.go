package mailchimp

import "fmt"

const (
	main_path   = "/lists/%s/members"
	member_path = main_path + "/%s"
	delete_path = member_path + "/actions/delete-permanent"
)

// Constants for member status
const (
	Subscribed    string = "subscribed"
	Unsubscribed  string = "unsubscribed"
	Cleaned       string = "cleaned"
	Pending       string = "pending"
	Transactional string = "transactional"
)

// Constants for HTTP methods
const (
	Post   string = "POST"
	Get    string = "GET"
	Put    string = "PUT"
	Patch  string = "PATCH"
	Delete string = "DELETE"
)

// MergeFields are the merge fields for a member
type MergeFields struct {
	FirstName      string `json:"FNAME,omitempty"`
	LastName       string `json:"LNAME,omitempty"`
	Programme      string `json:"MMERGE3,omitempty"`
	GraduationYear any    `json:"YEAR,omitempty"` // string if empty, int otherwise
}

// MemberRequest is the request body for adding or updating a member
type MemberRequest struct {
	Email       string      `json:"email_address"`
	Status      string      `json:"status,omitempty"`
	MergeFields MergeFields `json:"merge_fields,omitempty"`
}

// MemberResponse is the response body for retrieving, adding or updating a member
type MemberResponse struct {
	Id          string      `json:"id"`
	Email       string      `json:"email_address"`
	EmailId     string      `json:"unique_email_id"`
	ContactId   string      `json:"contact_id"`
	FullName    string      `json:"full_name"`
	Status      string      `json:"status"`
	MergeFields MergeFields `json:"merge_fields,omitempty"`
}

func (api *MailchimpAPI) GetMember(id *string) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Get, fmt.Sprintf(member_path, api.ListId, *id), nil, nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (api *MailchimpAPI) AddMember(request *MemberRequest) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Post, fmt.Sprintf(main_path, api.ListId), nil, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (api *MailchimpAPI) UpdateMember(id *string, request *MemberRequest) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Patch, fmt.Sprintf(member_path, api.ListId, *id), nil, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (api *MailchimpAPI) DeleteMember(id *string) error {
	err := api.Request(Post, fmt.Sprintf(delete_path, api.ListId, *id), nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
