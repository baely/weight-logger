package database

import (
	"context"
	"errors"
	"fmt"

	"firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/baely/weightloss-tracker/internal/util"
)

type Document struct {
	Title         string
	ActiveEnergy  float64
	RestingEnergy float64
	IntakeEnergy  float64
	Weight        float64
}

type TokenDocument struct {
	Token string
}

const (
	weightLogCollection = "weightlog"
	tokenCollection     = "token"
	tokenDocument       = "token"
)

func (d Document) InsertOrUpdate() error {
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: util.Project}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create a firebase app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to create a database client: %v\n", err)
	}
	defer client.Close()

	docRef := client.Collection("weightlog").Doc(d.Title)

	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			_, err = docRef.Set(ctx, d)
			if err != nil {
				return fmt.Errorf("failed to create new document: %v\n", err)
			}
		} else {
			return fmt.Errorf("failed to get document: %v\n", err)
		}
	} else {
		var firestoreDocument Document
		err = docSnapshot.DataTo(&firestoreDocument)
		if err != nil {
			return fmt.Errorf("failed to unmarshall database document: %v\n", err)
		}
		if d != firestoreDocument {
			_, err := docRef.Set(ctx, d)
			if err != nil {
				return fmt.Errorf("failed to update the document: %v\n", err)
			}
		}
	}

	return nil
}

func InsertOrUpdateDocuments(documents []Document) {
	for _, document := range documents {
		err := document.InsertOrUpdate()
		if err != nil {
			fmt.Printf("error saving document '%s': %v\n", document.Title, err)
		}
	}
}

func GetAllDocuments() ([]Document, error) {
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: util.Project}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create a firebase app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create a database client: %v\n", err)
	}
	defer client.Close()

	collection := client.Collection(weightLogCollection)

	iter := collection.Documents(ctx)

	docs := make([]Document, 0)

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		var document Document
		err = doc.DataTo(&document)
		if err != nil {
			return nil, err
		}

		docs = append(docs, document)
	}

	return docs, nil
}

func (d TokenDocument) InsertOrUpdate() error {
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: util.Project}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create a firebase app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to create a database client: %v\n", err)
	}
	defer client.Close()

	docRef := client.Collection(tokenCollection).Doc(tokenDocument)

	_, err = docRef.Set(ctx, d)
	if err != nil {
		return err
	}

	return nil
}

func GetToken() (TokenDocument, error) {
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: util.Project}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return TokenDocument{}, fmt.Errorf("failed to create a firebase app: %v\n", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return TokenDocument{}, fmt.Errorf("failed to create a database client: %v\n", err)
	}
	defer client.Close()

	docRef := client.Collection(tokenCollection).Doc(tokenDocument)
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		return TokenDocument{}, err
	}

	var t TokenDocument
	err = docSnapshot.DataTo(&t)
	if err != nil {
		return TokenDocument{}, err
	}

	return t, nil
}
