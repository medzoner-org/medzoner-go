# Publier un article

Créer `internal/resources/editorial/mon-article.md` :

```markdown
---
slug: mon-article
title: Un titre explicite
description: Une description pour les listes et moteurs de recherche.
deck: Une introduction à l’article.
author:
  name: Mehdi Youb
  url: https://www.medzoner.com/
category: go
tags: [go, open-source]
language: fr
image: /public/images/mon-article.png
image_alt: Description de l’image de partage.
status: draft
---

## Première section {#premiere-section}

Du texte, des **explications**, des liens et du code Markdown.
```

Ajouter l'image dans `public/images/`. Les slugs, rubriques et tags utilisent des
minuscules ASCII, chiffres et tirets. Les langues acceptées sont `fr` et `en`.
Les métadonnées inconnues, doublons de slug, statuts invalides et sources mal
formées font échouer le démarrage et les tests du catalogue embarqué.

Pour publier, passer `status` à `published` et ajouter un `published_at` réel en
RFC3339, avec fuseau horaire, par exemple `2026-10-25T14:00:00+01:00`.
Une date future planifie la visibilité : le processus compare la date à l'heure
courante à chaque requête. `updated_at` est facultatif et ne peut pas précéder
`published_at` ; ne le modifier que lors d'une réelle mise à jour.

Utiliser des blocs de code clôturés et des images Markdown avec texte alternatif.
Les ancres explicites `{#identifiant}` préservent les liens vers les sections.
Seuls les identifiants d'ancres sont autorisés dans les attributs Markdown :
les styles, classes et attributs interactifs sont supprimés. Un bloc de citation
`>` devient un encadré de note ; commencer ce bloc par `[!CTA]` dans un paragraphe
séparé crée un encadré d'appel à l'action et transforme ses liens isolés en boutons.
Le titre facultatif d'une image Markdown devient sa légende (`![Alt](url "Légende")`).
Le HTML brut est désactivé, y compris les iframes et scripts. Ne pas remettre
`WithUnsafe` pour contourner cette règle.

Valider avec `go test ./...`, `make lint`, `make build`, puis livrer via Git et la
CI. Aucun nouveau handler, template par article ou changement de DI n'est requis.
Les sources sont embarquées : il faut reconstruire et redéployer pour toute
modification. Le serveur n'offre pas de prévisualisation publique des brouillons.

## Points d'entrée

- `/articles` : catalogue publié, 10 résultats par page.
- `/articles?category=go&tag=open-source&page=1` : filtres combinés.
- `/articles/mon-article` : détail public ; 404 si absent, brouillon ou futur.
- `/feed.xml` : RSS 2.0 des 20 articles publiés les plus récents.
- `/sitemap.xml` : accueil, catalogue et URLs des articles publics.

La sélection des teasers de l'accueil et des raccourcis du header reste manuelle.
L'ajout d'un article le fait apparaître automatiquement dans le catalogue, le
sitemap et le RSS. Les brouillons ne sont pas un coffre-fort : le code source et
le binaire les contiennent. Ne pas y inclure de données confidentielles.
