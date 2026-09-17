package server_test

import "github.com/solanab/nsk/internal/client"

type stubForum struct {
	err      error
	who      *client.UserInfo
	user     *client.UserInfo
	cats     []client.Category
	list     *client.PostList
	post     *client.PostDetail
	pages    map[int]*client.PostDetail
	search   *client.SearchResult
	notes    []client.Notification
	lastPage int
	lastSlug string
	lastID   int
	lastQ    string
}

func newStub() *stubForum {
	list := new(client.PostList)
	list.Page = 1
	list.Posts = []client.PostSummary{{ID: 9, Title: titleHello, Author: nameAlice, URL: "u"}}

	post := new(client.PostDetail)
	post.ID = 9
	post.Title = titleHello
	post.Author = nameAlice
	post.Page = 1
	post.Pages = 1
	post.Floors = []client.Floor{{Number: 0, Author: nameAlice, Markdown: "hi"}}

	search := new(client.SearchResult)
	search.Query = queryVPS
	search.Page = 1
	search.Posts = list.Posts

	return &stubForum{
		who:    sampleUser(),
		user:   sampleUser(),
		cats:   []client.Category{{Slug: slugTech, Name: "技术"}},
		list:   list,
		post:   post,
		search: search,
		notes:  []client.Notification{{PostID: 9, URL: "u", Text: "n"}},
	}
}

func (stub *stubForum) WhoAmI() (*client.UserInfo, error) {
	return stub.who, stub.err
}

func (stub *stubForum) GetUser(id int) (*client.UserInfo, error) {
	stub.lastID = id

	return stub.user, stub.err
}

func (stub *stubForum) Categories() ([]client.Category, error) {
	return stub.cats, stub.err
}

func (stub *stubForum) LatestPosts(page int) (*client.PostList, error) {
	stub.lastPage = page
	stub.lastSlug = ""

	return stub.list, stub.err
}

func (stub *stubForum) CategoryPosts(slug string, page int) (*client.PostList, error) {
	stub.lastSlug = slug
	stub.lastPage = page

	return stub.list, stub.err
}

func (stub *stubForum) GetPost(postID, page int) (*client.PostDetail, error) {
	stub.lastID = postID
	stub.lastPage = page

	if stub.err != nil {
		return nil, stub.err
	}

	if stub.pages != nil {
		if detail, ok := stub.pages[page]; ok {
			return detail, nil
		}

		empty := new(client.PostDetail)
		empty.ID = postID
		empty.Page = page

		return empty, nil
	}

	return stub.post, nil
}

func (stub *stubForum) GetPostAll(postID int) (*client.PostDetail, error) {
	return stub.GetPost(postID, 1)
}

func (*stubForum) FormatPost(detail *client.PostDetail) string {
	return client.FormatPost(detail)
}

func (stub *stubForum) Search(query string, page int) (*client.SearchResult, error) {
	stub.lastQ = query
	stub.lastPage = page

	return stub.search, stub.err
}

func (stub *stubForum) Notifications() ([]client.Notification, error) {
	return stub.notes, stub.err
}

var _ client.Forum = (*stubForum)(nil)
