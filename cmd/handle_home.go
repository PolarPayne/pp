package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
)

var tmpl = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link rel="shortcut icon" href="/logo">

	{{ if .NotLoggedIn }}
	<title>Private Podcast</title>
	{{ else }}
	<title>Private Podcast - {{ .Name }}</title>
	{{ end }}

	<style>
		body {
			background: #ccc;
		}

		audio {
			width: 100%;
		}

		.content {
			background: white;
			max-width: 80em;
			margin: 1rem auto;
			padding: 1rem;
			border: 2px solid black;
		}

		.description {
			margin-bottom: 4rem;
		}

		.alert {
			font-weight: 900;
			font-size: 110%;
		}

		.url {
			font-family: monospace;
			overflow-x: auto;
		}
	</style>
</head>
<body>
	<div class="content">

{{ if .NotLoggedIn }}
	<a href="/?action=login">login</a>
	{{ else }}
	<a href="/?action=logout">logout</a>
{{ end }}

{{ if not .NotLoggedIn }}
	<h1>{{ .Name }}</h1>
	<p class="description">{{ .Description }}</p>

	<p>The following URL is your private podcast feed. <span class="alert">DO NOT SHARE IT WITH ANYONE.</span> We track all requests.</p>
	<p class="url copytext">{{ .FeedURL }}</p>
	<button onclick="copyText('.copytext')">copy link</button>
	<p>This URL should work with pretty much any podcast application that supports custom URLs (at least <a href="https://www.videolan.org/vlc/">VLC</a> and <a href="https://overcast.fm/">Overcast</a> are known to work), just <span class="alert">DON'T SHARE IT</span>.</p>

	<hr>

	<h2>Episodes</h2>

	{{ range .Podcasts }}
	<div class="podcast">
		<h3>{{ .Title }} ({{ .Published }})</h3>
		<audio controls src="{{ .URL }}">Your browser does not support the <code>audio</code> element.</audio>
	</div>
	{{ end }}

{{ end }}

	<hr>

	<p>If you're having technical problems please
	{{ if .Help }}
		{{ .Help }}
	{{ else }}
		see <a href="//github.com/polarpayne/pp">github.com/polarpayne/pp</a>.</p>
	{{ end }}
	</p>

	</div>

	<script>
	function copyText(el) {
		var copyTextArea = document.querySelector(el);

		try {
			navigator.clipboard.writeText(copyTextArea.innerText)
			.then(() => console.log("Copied text succesfully!"))
			.catch(() => alert("Unable to copy to clipboard. :("));
		} catch (err) {
			alert("Unable to copy to clipboard. :(\nYou'll need to manually copy the URL.");
		}
	}
	</script>
</body>
`

func (s *server) handleHome() http.HandlerFunc {
	tmplCompiled := template.Must(template.New("home").Parse(tmpl))

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("podcast_session")
		if err != nil && err != http.ErrNoCookie {
			s.handleError(w, r, err)
			return
		}

		sessionCookieNotSet := err == http.ErrNoCookie
		var secret string
		if !sessionCookieNotSet {
			secret = c.Value
		}
		action := r.URL.Query().Get("action")

		switch action {
		case "login":
			http.Redirect(w, r, "/auth", 302)
			return

		case "logout":
			http.SetCookie(w, &http.Cookie{Name: "podcast_session", MaxAge: -1})
			http.Redirect(w, r, "/", 302)
			return
		}

		q := url.Values{}
		q.Set("s", secret)
		feedURL := s.baseURL + "/feed?" + q.Encode()

		type p struct {
			Title, URL string
			Published  string
		}
		podcasts := make([]p, 0)
		for _, podcast := range s.getPodcasts() {
			pd := podcast.Details()

			q = url.Values{}
			q.Set("s", secret)
			q.Set("n", pd.Key)
			pURL := s.baseURL + "/podcast?" + q.Encode()
			podcasts = append(podcasts, p{pd.Title, pURL, pd.Published.Format("2006-01-02")})
		}

		err = tmplCompiled.Execute(w, struct {
			Secret, FeedURL         string
			NotLoggedIn             bool
			Name, Description, Help string
			Podcasts                []p
		}{secret, feedURL, sessionCookieNotSet, s.name, s.description, s.helpText, podcasts})
		if err != nil {
			log.Printf("failed to render home: %v", err)
		}
	}
}
