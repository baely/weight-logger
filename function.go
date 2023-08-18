package weightloss_tracker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"google.golang.org/protobuf/proto"

	"github.com/baely/weightloss-tracker/internal/database"
	"github.com/baely/weightloss-tracker/internal/integrations/gcs"
	"github.com/baely/weightloss-tracker/internal/util"
	"github.com/baely/weightloss-tracker/internal/util/image"
)

func EventDocumentToDocument(eventDoc *firestoredata.Document) database.Document {
	fields := eventDoc.GetFields()

	doc := database.Document{
		Title:         fields["Title"].GetStringValue(),
		ActiveEnergy:  fields["ActiveEnergy"].GetDoubleValue(),
		RestingEnergy: fields["RestingEnergy"].GetDoubleValue(),
		IntakeEnergy:  fields["IntakeEnergy"].GetDoubleValue(),
		Weight:        fields["Weight"].GetDoubleValue(),
	}

	return doc
}

func GenerateProgressImage(ctx context.Context, event event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(event.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	value := data.GetValue()
	if value == nil {
		return nil
	}

	doc := EventDocumentToDocument(value)

	img, err := image.Generate(doc)
	if err != nil {
		fmt.Println("error gen image:", err)
		return err
	}
	b := bytes.NewReader(img)

	filename := fmt.Sprintf(image.FilenameFormat, doc.Title)
	err = gcs.UploadFile(util.ResourceBucket, filename, b)
	if err != nil {
		fmt.Println("error saving doc:", err)
		return err
	}

	return nil
}
