package main

import (
	"context"
	"fmt"
	robotomail "github.com/robotomail/robotomail-go"
	"log"
)

func main() {
	mail, err := robotomail.NewClient(robotomail.Options{}) // reads ROBOTOMAIL_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	result, err := mail.ListMailboxes(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, box := range result.Mailboxes {
		fmt.Println(box.FullAddress)
	}
}
