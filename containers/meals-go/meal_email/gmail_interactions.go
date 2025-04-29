package meal_email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := os.Getenv("TOKEN_LOCATION")
	if tokFile == "" {
		log.Fatal("TOKEN_LOCATION not set.")
	}

	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		err := saveToken(tokFile, tok)
		if err != nil {
			return nil, fmt.Errorf("unable to save token: %v", err)
		}
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) error {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache ouath token: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}

	return nil
}

// Define your own type that wraps *gmail.Service
type GmailService struct {
	Service *gmail.Service
}

func AuthenticateGmail() (GmailService, error) {
	ctx := context.Background()
	credentialsLocation := os.Getenv("CREDENTIALS_LOCATION")
	if credentialsLocation == "" {
		log.Fatal("CREDENTIALS_LOCATION not set.")
	}

	b, err := os.ReadFile(credentialsLocation)
	if err != nil {
		log.Fatalf("Unable to read client secret file %s: %v", credentialsLocation, err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(config)
	if err != nil {
		log.Fatalf("Unable to get Gmail client: %v", err)
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to spawn new service Gmail client: %v", err)
	}

	return GmailService{Service: srv}, nil
}

func (gs *GmailService) SendEmail(from string, to []string, subject, body string) error {
	header := make(map[string]string)
	header["From"] = from
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	var message strings.Builder
	for k, v := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + body)

	rawMessage := base64.URLEncoding.EncodeToString([]byte(message.String()))
	gmailMessage := &gmail.Message{
		Raw: rawMessage,
	}

	_, err := gs.Service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("unable to send email: %v", err)
	}
	return nil
}
