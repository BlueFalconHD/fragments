package main

import (
	"os"
	"path/filepath"
)

const defaultConfig = `# Welcome to your site's configuration file!'
# This file is in YAML format. If you are unfamiliar with YAML, please see https://yaml.org/start.html

# The path to the directory containing your site's fragments, relative to this file
# Fragments are reusable pieces of content that can be included in your page.
fragments: fragment

# The path to the directory containing your site's pages, relative to this file
# Pages are turned directly into HTML files, so the structure of these determines the structure of the pages in your site.
pages: page

# The path to the directory containing extra files for your site
# Any CSS, JS, or assets should be included here to be copied to the output directory
include: include

# The final output directory of your site
# If you don't want this to be committed to git, make sure to add it to your .gitignore file
build: build
`

const defaultIndexPage = `this:setTemplate("page")

~~~

<h1 class="title">Welcome to my site!</h1>
<p class="description">This is a simple site built with <a href="https://github.com/bluefalconhd/fragments">Fragments</a>.</p>

<h1>Recent Blogposts</h1>
@{blogposts}`

const defaultExamplePost = `this:setTemplate("post")

this:setSharedMeta {
    postTitle = "Example Post",
    postDescription = "This is an example post.",
    postDate = os.date("%Y-%m-%d"),
    author = "Lorem Ipsum"
}

~~~

# Sexta populus coniugium flabat socio

## Nato Erectheas pudet

Lorem markdownum en mihi figuram. Emittere honore! Prosunt dedisset signans
dominaeque nuda; atra ardua nomina, tu hoc? Sum tibi equus quid tauros; frustra
inridet: curis male idem pedes nitidaeque esse!

- Senem quam molitur parva
- Saevitiam temone
- Effreno cum non magnae
- Dum Tegeaea carmine
- Genitalia parva insequitur credentes venisse fessa prodibant

## Est Iove sub similis latet

Aurora tamen et taedia saecula genetrici dixit et cupiere forma serpere. Gradere
pariter felix soporem velari. *Surgit quod exequialia* praeceps aureus ad unde
flenti, inanes citharam pectora indignatus.

` + "```" + `
hardDriver += sdramInput;
icsPharmingPhreaking.motherboard += snippet_smart - 3 + vlb;
var navigation = sector_case_boolean(null_utility_bluetooth + hashtagJpeg,
        metafileWeb(restore));
` + "```" + `

## Fugit et inritat

Unus ubi et removete videt dea domum ab tauros artem. Cui tempora casus.

## Grandine Tartara

Ora a sepulcro obstitit *loquendo iuvenes aditumque* et bubo! Pensas inde factis
dare *amantem* agmenque nisi, ad conpagibus rebus eadem movet. Manibus sua
rediit tollit frendens et lubrica Aeacus; non sors et atque vocis arma. Furtoque
tempora, voce Tarpeia procubuit sanguine: anno cinctaeque dirum, sic
[dabatur](http://www.ab-commissaque.com/telumiunguntur) anilia Tyrrhenia adsere.
*Aut vos ligno* inpia laticem ingredior cavus.

1. Sparsaque ille
2. Nec vulnera ignes
3. Latissima sustinet

Essemus potentia Caucasus rasilis. Solebat et dolor *et urebat causam* pallida
si meritis plaustri. Spirarunt amaris stravit quoque terrigenae; semina est
aequoreas saxaque adest mensura Latialis.`

const defaultPageFragment = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fragments Site</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    @{nav}

    <main>
        ${CONTENT}
    </main>

    @{footer}
</body>`

const defaultPostFragment = `this:setTemplate("page")

-- We have the following meta available to us for a standard post
-- postTitle
-- postDate
-- postDescription
-- author

~~~
<article>
    <h1>${postTitle}</h1>
    @{markdown[[${CONTENT}]]}
</article>`

const defaultNavFragment = `<header class="nav">
    <div class="nav-logo">
        <a href="/index.html" class="unstyled-link">
           <span class="logo-text">⌘</span> My Site
        </a>
    </div>
</header>`

// Function to setup files in a certain directory
func setupFiles(dir string) error {
	// Ensure root directory exists
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Create required subdirectories
	if err := os.MkdirAll(filepath.Join(dir, "page", "posts"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "fragment"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "include"), os.ModePerm); err != nil {
		return err
	}

	// Create the config file
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(defaultConfig), os.ModePerm); err != nil {
		return err
	}

	// Write default fragments
	if err := os.WriteFile(filepath.Join(dir, "fragment", "page.frag"), []byte(defaultPageFragment), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "fragment", "post.frag"), []byte(defaultPostFragment), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "fragment", "nav.frag"), []byte(defaultNavFragment), os.ModePerm); err != nil {
		return err
	}
	footer := "<footer class=\"footer\">\n    <p>&copy; 2024 Your Name</p>\n    <p>Powered by <a href=\"https://github.com/bluefalconhd/fragments\">Fragments</a></p>\n</footer>\n"
	if err := os.WriteFile(filepath.Join(dir, "fragment", "footer.frag"), []byte(footer), os.ModePerm); err != nil {
		return err
	}

	// Write example pages
	if err := os.WriteFile(filepath.Join(dir, "page", "index.frag"), []byte(defaultIndexPage), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "page", "posts", "example.frag"), []byte(defaultExamplePost), os.ModePerm); err != nil {
		return err
	}

	// Write a default stylesheet
	css := "body { font-family: system-ui; margin: 0; padding: 0; background-color: #1e1e2e; color: #cdd6f4; }\n"
	if err := os.WriteFile(filepath.Join(dir, "include", "style.css"), []byte(css), os.ModePerm); err != nil {
		return err
	}

	return nil
}
