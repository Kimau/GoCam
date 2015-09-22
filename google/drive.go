package google

import (
	"fmt"
	"io"

	drive "google.golang.org/api/drive/v2"
)

const (
	mimeDoc string = "application/vnd.google-apps.document"
)

// AllRevisions fetches all revisions for a given file
func AllRevisions(fileId string) ([]*drive.Revision, error) {
	<-driveThrottle // rate Limit
	r, err := drvSvc.Revisions.List(fileId).Do()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	return r.Items, nil
}

// AllFiles fetches and displays all files
func AllFiles(query string, pageNum chan int) ([]*drive.File, error) {
	var fs []*drive.File
	pageToken := ""
	count := 0
	for {
		count = count + 1

		q := drvSvc.Files.List()
		q.Spaces("drive") // Only get drive (not 'appDataFolder' 'photos')
		q.Q(query)

		// If we have a pageToken set, apply it to the query
		if pageToken != "" {
			q = q.PageToken(pageToken)
		}

		pageNum <- count
		<-driveThrottle // rate Limit
		r, err := q.Do()
		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
			return fs, err
		}
		fs = append(fs, r.Items...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	pageNum <- -1
	return fs, nil
}

func InsertFile(title string, description string, parentId string, mimeType string, data io.Reader) (*drive.File, error) {

	f := &drive.File{Title: title, Description: description, MimeType: mimeType}
	if parentId != "" {
		p := &drive.ParentReference{Id: parentId}
		f.Parents = []*drive.ParentReference{p}
	}

	r, err := drvSvc.Files.Insert(f).Media(data).Ocr(false).Convert(false).Do()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	return r, nil
}
