package models

import "slices"

type Project struct {
	Slug            string
	Name            string
	Tagline         string
	Summary         string
	RepoURL         string
	LiveURL         string
	LastUpdate      string
	FirstCommitDate string
	Notes           string
	Tech            []string
	Highlights      []string
	Paragraphs      []string
	Screenshots     []ProjectScreenshot
}

type ProjectScreenshot struct {
	Title string
	Path  string
	Alt   string
	Note  string
}

func Projects() []Project {
	projects := []Project{
		{
			Slug:            "dined",
			Name:            "Dined",
			Tagline:         "Proof that nobody actually agreed on dinner.",
			Summary:         "A private family restaurant memory ledger for remembering where everyone ate, who picked it, and how the table rated it.",
			RepoURL:         "https://github.com/bitofbytes-io/dined",
			LiveURL:         "https://dined.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2026-05-10",
			Notes:           "Most recent work added public self-hosting guidance, documented safer local secret handling, and kept auth and dining flow observability coverage current.",
			Tech: []string{
				"Go",
				"HTMX",
				"Tailwind-style CSS",
				"PostgreSQL",
				"Goose",
				"Google Places API",
				"Docker Swarm",
			},
			Highlights: []string{
				"Logs restaurant visits with picker, attendee ratings, price level, tags, and notes so the family can remember what happened next time.",
				"Keeps public readonly browsing separate from authenticated write controls for adding visits and managing restaurant metadata.",
				"Wraps the workflow in a retro diner identity with booth, ticket-pad, and trophy-case screens.",
			},
			Paragraphs: []string{
				"Dined is a private family restaurant memory ledger for answering the question: have we eaten here before, who picked it, and did we like it? It tracks restaurants, dining visits, pickers, participant ratings, tags, and notes without trying to become a public review network.",
				"The app is built around a mobile-first logging workflow with Google Places lookup, a chronological public dine history, restaurant detail pages, and playful trophy-case stats. The visual direction leans into an authentic retro diner booth and jukebox feel so the product has a distinct personality around a simple family habit.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Booth home",
					Path:  "/static/projects/dined/booth-home.png",
					Alt:   "Dined home screen with a retro diner booth scene and recent dine cards.",
					Note:  "Captured from the local Dined memory preview.",
				},
				{
					Title: "Trophy case",
					Path:  "/static/projects/dined/trophy-case.png",
					Alt:   "Dined trophy case screen with record-style stats, a dining map, and top restaurant rankings.",
					Note:  "Captured from the live Dined trophy case.",
				},
			},
		},
		{
			Slug:            "permitpal",
			Name:            "PermitPal",
			Tagline:         "Your permit pal-less yelling, more tracking.",
			Summary:         "A focused dashboard for tracking North Carolina learner permit exam readiness and study progress.",
			RepoURL:         "https://github.com/bitofbytes-io/permitpal",
			LiveURL:         "https://permitpal.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2026-05-01",
			Notes:           "Most recent work added public self-hosting guidance, documented safer local secret handling, and kept auth, dashboard, request, and security observability coverage current.",
			Tech: []string{
				"Go",
				"HTMX",
				"Templ",
				"Tailwind CSS",
				"PostgreSQL",
				"Goose",
				"Docker",
			},
			Highlights: []string{
				"Shows permit-readiness progress at a glance so studying feels measurable instead of vague.",
				"Breaks the exam material into trackable practice areas and completion states.",
				"Keeps the workflow narrow: check progress, see weak spots, and decide what to study next.",
			},
			Paragraphs: []string{
				"PermitPal is for turning learner permit prep into a simple progress dashboard. It gives a driver-in-training a place to see what has been reviewed, what still needs attention, and whether the next study session should focus on signs, rules, or weak areas.",
				"The app is intentionally small and task-focused. Instead of becoming a general learning platform, it keeps the experience centered on readiness: open the dashboard, review the current state, and make the next study decision quickly.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Local dashboard",
					Path:  "/static/projects/permitpal/dashboard.png",
					Alt:   "PermitPal local dashboard with progress meters and a skill mastery checklist.",
					Note:  "Captured from the local PermitPal preview at localhost:4600.",
				},
				{
					Title: "Mobile dashboard",
					Path:  "/static/projects/permitpal/dashboard-mobile.png",
					Alt:   "PermitPal dashboard captured at a mobile viewport.",
					Note:  "Captured from the local in-memory workflow.",
				},
			},
		},
		{
			Slug:            "learnd",
			Name:            "Learn'd",
			Tagline:         "A personal learning journal for capturing and reviewing resources.",
			Summary:         "A personal learning journal for saving useful training material and revisiting it over time.",
			RepoURL:         "https://github.com/bitofbytes-io/learnd",
			LiveURL:         "https://learnd.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2026-01-02",
			Notes:           "Most recent work added public self-hosting guidance, protected local setup secrets, removed outdated AI-agent guides, and refreshed sanitized auth and learning flow observability notes.",
			Tech: []string{
				"Go",
				"Templ",
				"Tailwind CSS",
				"PostgreSQL",
				"Goose",
				"JavaScript",
				"Playwright",
			},
			Highlights: []string{
				"Saves useful courses, articles, videos, and references into one private review space.",
				"Helps turn one-off learning links into a backlog that can be revisited later.",
				"Supports a personal workflow for tracking what was worth keeping and what should be reviewed next.",
			},
			Paragraphs: []string{
				"Learn'd is for capturing training material before it disappears into bookmarks, chat threads, or browser history. It gives learning resources a dedicated home so useful material can be saved, organized, and revisited when there is time to go deeper.",
				"The app is shaped around a private learning habit: collect the resource, remember why it mattered, and make it easier to come back to the right item later.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Resource journal",
					Path:  "/static/projects/learnd/dashboard.png",
					Alt:   "Learn'd resource journal with a capture form and recent learning entries.",
					Note:  "Captured from the authenticated Learn'd live site.",
				},
				{
					Title: "Mobile capture",
					Path:  "/static/projects/learnd/mobile.png",
					Alt:   "Learn'd mobile view showing the resource capture form and recent entries.",
					Note:  "Captured from the authenticated mobile workflow.",
				},
			},
		},
		{
			Slug:            "dejaview",
			Name:            "DejaView",
			Tagline:         "A family movie-night tracker for remembering picks, groups, and stats.",
			Summary:         "A movie tracker for organizing watched films, family viewing lists, and collection details.",
			RepoURL:         "https://github.com/bitofbytes-io/dejaview",
			LiveURL:         "https://dejaview.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2026-01-04",
			Notes:           "Most recent work added public self-hosting prerequisites, removed outdated AI-agent guides, and kept auth, movie, entry, and rating flow observability notes current.",
			Tech: []string{
				"Go",
				"Templ",
				"Tailwind CSS",
				"PostgreSQL",
				"Goose",
				"TMDB API",
			},
			Highlights: []string{
				"Organizes movie lists around repeat viewing, family picks, and collection context.",
				"Turns movie tracking into a browsable catalogue instead of a plain spreadsheet.",
				"Includes a stats page with funny trophies for family movie-night bragging rights.",
			},
			Paragraphs: []string{
				"DejaView is for keeping a personal movie collection understandable at a glance. It gives watched films and family movie-night lists a place to live, making it easier to remember what has been watched, what belongs together, and what might be worth revisiting.",
				"The experience is built around browsing and recognition: poster-forward entries, grouped lists, and enough detail to make the collection feel useful without turning movie tracking into busywork.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Movie collection",
					Path:  "/static/projects/dejaview/movie-collection-posters.png",
					Alt:   "DejaView public movie collection screen grouped by family movie night lists.",
					Note:  "Captured from the public DejaView site.",
				},
				{
					Title: "Stats view",
					Path:  "/static/projects/dejaview/stats.png",
					Alt:   "DejaView authenticated stats page with family movie-night awards and pick metrics.",
					Note:  "Captured from the authenticated DejaView stats view.",
				},
			},
		},
		{
			Slug:            "bitofbytes",
			Name:            "BitOfBytes",
			Tagline:         "The personal landing page and project portfolio for Daniel Waters.",
			Summary:         "A personal site for presenting resume details, contact links, and selected project work.",
			RepoURL:         "https://github.com/bitofbytes-io/bitofbytes",
			LiveURL:         "https://www.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2024-06-29",
			Notes:           "Most recent work refreshed project recency notes, added public self-hosting guidance, and clarified safe local environment setup for the portfolio.",
			Tech: []string{
				"Go",
				"HTML templates",
				"Tailwind CSS",
				"Gorilla CSRF",
				"Docker",
			},
			Highlights: []string{
				"Presents resume details, contact links, and selected personal projects in one focused site.",
				"Gives each project room for screenshots, purpose, highlights, and links to the live app or code.",
				"Removes older blog and utility pages so the site stays centered on professional context and current work.",
			},
			Paragraphs: []string{
				"BitOfBytes is the portfolio site for presenting who Daniel is, how to get in touch, and what personal projects are worth looking at. The project pages are meant to give each app enough context that a visitor can understand its purpose before opening the repo.",
				"The current version trims the site down to the useful surface area: identity, contact, resume, and a small catalogue of project pages with screenshots and plain-language notes.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Home page",
					Path:  "/static/projects/bitofbytes/home.png",
					Alt:   "The BitOfBytes home page using the Field Notebook design system.",
					Note:  "Captured from the rebuilt local BitOfBytes home page.",
				},
				{
					Title: "Project index",
					Path:  "/static/projects/bitofbytes/projects.png",
					Alt:   "The rebuilt BitOfBytes projects index page.",
					Note:  "Captured from the rebuilt local project index.",
				},
			},
		},
		{
			Slug:            "anthology",
			Name:            "Anthology",
			Tagline:         "A two-tier catalogue for personal books, games, movies, and music.",
			Summary:         "A personal media catalogue for organizing books, games, movies, music, and where they live.",
			RepoURL:         "https://github.com/bitofbytes-io/anthology",
			LiveURL:         "https://anthology.bitofbytes.io",
			LastUpdate:      "July 10, 2026",
			FirstCommitDate: "2025-10-30",
			Notes:           "Most recent work added public self-hosting and safer container quickstart guidance, clarified production cookie mode, removed outdated AI-agent guides, and refreshed workflow route notes.",
			Tech: []string{
				"Go",
				"Chi",
				"Angular",
				"Angular Material",
				"TypeScript",
				"PostgreSQL",
				"Google OAuth",
				"Google Books API",
			},
			Highlights: []string{
				"Catalogues books, games, movies, and music across one personal media inventory.",
				"Models where items physically live so shelves and storage locations are part of the collection.",
				"Supports import and lookup workflows for growing the catalogue without hand-entering every detail.",
			},
			Paragraphs: []string{
				"Anthology is for managing a personal media collection that spans formats: books, games, movies, music, and the places they are stored. It treats the catalogue as both a list of items and a map of where those items live.",
				"The app is aimed at collection maintenance. It helps add items, enrich them with useful metadata, import larger batches, and keep shelves or storage areas understandable over time.",
			},
			Screenshots: []ProjectScreenshot{
				{
					Title: "Catalogue dashboard",
					Path:  "/static/projects/anthology/catalogue.png",
					Alt:   "Anthology catalogue table with filters and media items.",
					Note:  "Captured from the authenticated Anthology live site.",
				},
				{
					Title: "Shelves workflow",
					Path:  "/static/projects/anthology/shelves.png",
					Alt:   "Anthology shelves screen showing storage locations and item counts.",
					Note:  "Captured from the authenticated shelves workflow.",
				},
			},
		},
	}

	slices.SortFunc(projects, func(a, b Project) int {
		switch {
		case a.FirstCommitDate > b.FirstCommitDate:
			return -1
		case a.FirstCommitDate < b.FirstCommitDate:
			return 1
		default:
			return 0
		}
	})

	return slices.Clone(projects)
}
