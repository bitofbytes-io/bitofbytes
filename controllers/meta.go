package controllers

import (
	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/views"
)

const siteName = "BitOfBytes"

func homeMeta() views.Meta {
	const description = "Daniel Waters is a senior software engineer in Raleigh, NC. He builds live video and real-time collaboration software at Cisco, and small apps for everyday life at home."
	return views.Meta{
		Title:       siteName + " | Daniel Waters",
		Description: description,
		Path:        "/",
		JSONLD: map[string]any{
			"@context":    "https://schema.org",
			"@type":       "Person",
			"name":        "Daniel Waters",
			"url":         views.SiteURL + "/",
			"jobTitle":    "Senior Software Engineer",
			"description": description,
			"address": map[string]any{
				"@type":           "PostalAddress",
				"addressLocality": "Raleigh",
				"addressRegion":   "NC",
			},
			"sameAs": []string{
				"https://github.com/bitofbytes-io",
				"https://www.linkedin.com/in/daniel-waters/",
			},
		},
	}
}

func projectsMeta() views.Meta {
	return views.Meta{
		Title:       "Projects | " + siteName,
		Description: "Side projects by Daniel Waters: small, private apps for everyday life, each with screenshots and what changed most recently.",
		Path:        "/projects",
	}
}

func updatesMeta() views.Meta {
	return views.Meta{
		Title:       "Updates | " + siteName,
		Description: "Every update to Daniel Waters's side projects, newest first, also available as an Atom feed.",
		Path:        "/updates",
	}
}

func projectMeta(project models.Project) views.Meta {
	image := project.Thumbnail().Path
	jsonld := map[string]any{
		"@context":            "https://schema.org",
		"@type":               "SoftwareApplication",
		"name":                project.Name,
		"description":         project.Tagline,
		"url":                 views.SiteURL + "/projects/" + project.Slug,
		"applicationCategory": "WebApplication",
		"author": map[string]any{
			"@type": "Person",
			"name":  "Daniel Waters",
			"url":   views.SiteURL + "/",
		},
	}
	if image != "" {
		jsonld["image"] = views.SiteURL + image
	}
	if updated := project.UpdatedOn(); !updated.IsZero() {
		jsonld["dateModified"] = updated.Format(sitemapDateLayout)
	}
	return views.Meta{
		Title:       project.Name + " | " + siteName,
		Description: project.Tagline,
		Path:        "/projects/" + project.Slug,
		Image:       image,
		JSONLD:      jsonld,
	}
}
