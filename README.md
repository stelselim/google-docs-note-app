# 📝 Save Notes to Google Docs with Images

This project is a simple Go-based server that lets you save rich notes — including text, tags, and images — directly to a Google Docs document using the Google Drive and Docs APIs.

It’s designed to demonstrate OAuth 2.0 integration, working with multipart form uploads, and structured note-saving into a single shared document.

<br>

## Features

- [X] Google OAuth 2.0 Authentication
- [X] Upload images to Google Drive
- [X] Save notes (title, tags, text, and images) to a **single Google Docs file**
- [X] Insert each new note at the top of the document
- [X] Two REST API endpoints:
  - `POST /note`: Add a new note with optional images
  - `GET /notes`: Retrieve all notes in structured JSON format
- [X] Lightweight: No external deployment required — runs locally

<br>

## How It Works

1. **Authentication**: Uses OAuth 2.0 to authenticate and authorize access to Google Docs & Drive APIs.
2. **Image Upload**: Uploaded images are saved to Google Drive, and their URLs are retrieved.
3. **Note Formatting**: A note containing the title, tags, timestamp, and content (with inline images) is created.
4. **Google Docs Update**: The formatted note is inserted at the beginning of a predefined Google Doc (ID is set via an environment variable).

<br>

## Environment Variables

Set the following environment variable before running the app:

```bash
GOOGLE_DOCS_ID=your_google_docs_id_here
```

<br>

## Running Locally

1.	Clone the repo
2.	Set up your Google Cloud project (enable Drive & Docs APIs)
3.  Authenticate with Google (locally). In this project, you can use `gcloud auth application-default login` to set up Application Default Credentials.
4.	Set GOOGLE_DOCS_ID in your environment
5.	Run the server:

```bash
go run main.go
```

<br>

## Example Note at Google Docs

### Travel Plans in 2025 
Create Date: 2:08:38 PM, 5/20/2025 

Day 1: Start at Buckingham Palace, stroll St. James’s Park, and visit Westminster Abbey. See Big Ben and the Houses of Parliament, then head to the South Bank for lunch and a ride on the London Eye. Walk along the Thames to Tower Bridge, passing the Globe and Tate Modern. End with a sunset Thames cruise and riverside dinner.

Day 2: Visit the British Museum, then explore the Tower of London and Tower Bridge. Enjoy dinner and a show in the West End.

Day 3: Wander Notting Hill, relax in Hyde Park, visit Kensington museums, shop at Harrods, and dine in South Kensington.


<img src="https://www.studying-in-uk.org/wp-content/uploads/2019/05/study-in-london-1068x641.jpg" width=300>


<br>

## API Overview

### `POST /note`
Add a new note with optional tags and images.

#### Request (multipart/form-data)
- `title`: string  
- `note`: string (main text content)  
- `tags`: string
- `images[]`: file(s)

### `GET /notes`
Retrieve all notes in structured format.

#### Response
```json
{
  "success": true,
  "message": "2 note(s) found.",
  "data": [
    {
      "title": "New Note Without Image",
      "note": "- Complete your Go Project",
      "tags": "go,coding",
      "date": "06:27:40 PM, 2025-05-24"
    },
    {
      "title": "New Homework",
      "note": "- Read Unit 1 Words.\n- Do exercise 1.2 at the Lecture Book.\n",
      "tags": "homework,dutch,study",
      "date": "01:30:09 AM, 2025-05-23"
    }
  ]
}
```

<br>

## Tech Stack

- Language: **Go**
- API: `net/http`, `mime/multipart`, and Google SDKs
- Services Used: 
  - Google Drive API (image upload)
  - Google Docs API (note saving)
