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

// MemberRequest is the request body for adding or updating a member
type MemberRequest struct {
	Email  string `json:"email_address"`
	Status string `json:"status"`
}

// MemberResponse is the response body for adding or updating a member
type MemberResponse struct {
	Id        string `json:"id"`
	Email     string `json:"email_address"`
	EmailId   string `json:"unique_email_id"`
	ContactId string `json:"contact_id"`
	FullName  string `json:"full_name"`
	Status    string `json:"status"`
}

func GetMember(listId, id *string, api *MailchimpAPI) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Get, fmt.Sprintf(member_path, *listId, *id), nil, nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func AddMember(listId *string, request *MemberRequest, api *MailchimpAPI) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Post, fmt.Sprintf(main_path, *listId), nil, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func UpdateMember(listId, id *string, request *MemberRequest, api *MailchimpAPI) (*MemberResponse, error) {
	response := &MemberResponse{}

	err := api.Request(Patch, fmt.Sprintf(member_path, *listId, *id), nil, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func DeleteMember(listId, id *string, api *MailchimpAPI) error {
	err := api.Request(Post, fmt.Sprintf(delete_path, *listId, *id), nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
