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

## Suivre un projet qui évolue

`doc/editorial_sources.json` relie un article à un dépôt, une branche, un commit
**réellement relu** et des chemins pertinents. L'article doit contenir un lien
vers ce commit complet (`https://github.com/owner/repo/tree/<sha>`). Ne pas ajouter
ces champs au front matter : les métadonnées inconnues sont volontairement
refusées par le catalogue.

La veille `Editorial source watch` compare cette référence à la branche amont
chaque lundi à 07:00 UTC et peut être lancée manuellement dans Actions. Elle
ouvre ou actualise une issue de revue par article si des fichiers surveillés ont
changé. Les renommages, historiques divergents et limites de comparaison GitHub
(300 fichiers) sont traités de façon conservatrice. Les erreurs API font échouer
la vérification, elles ne signifient pas « à jour ». Les exécutions de publication
sont sérialisées ; une issue inchangée n'est pas réécrite à chaque passage.
Le bot ne gère que les issues créées par `github-actions[bot]`, portant le titre
attendu et le marqueur de l'article. Une copie du marqueur dans une issue tierce
ne lui donne pas le droit de l'écraser. Conserver le titre d'une issue gérée et
placer les notes de revue en commentaires : son corps est actualisé par le bot.

Cette veille **ne met pas la prose à jour automatiquement**. Elle ne lance aucun
modèle IA, n'exécute aucun code amont, n'importe aucun fichier source dans le site,
n'avance pas le commit relu et ne publie rien. Le token Actions a seulement la
lecture des contenus et l'écriture des issues pour le job de veille. Les PR ne
lancent que les tests locaux, sans accès d'écriture aux issues. Aucun PAT
inter-dépôts supplémentaire n'est nécessaire pour un amont public.

Après fusion de cette configuration sur la branche par défaut, avec Actions
activé, la planification devient disponible. GitHub peut retarder un horaire ou
désactiver les schedules d'un dépôt public inactif : le déclenchement manuel est
le recours, pas une garantie de fraîcheur permanente.

```sh
# Python 3.10+ et GitHub CLI authentifié. Lecture seule par défaut.
python3 scripts/editorial_source_watch.py

# Tests sans API GitHub ni création d'issue.
python3 -m unittest discover -s scripts/tests -p 'test_editorial_source_watch.py' -v

# Déclenchement manuel du workflow, une fois fusionné sur la branche par défaut.
gh workflow run editorial-source-watch.yml --repo medzoner-org/medzoner-go
```

Pour traiter une notification :

1. relire le diff amont et vérifier contrats, limites, tests et preuves matérielles ;
2. corriger l'article, les schémas et ses tests de rendu si nécessaire ;
3. préserver `published_at` et modifier `updated_at` pour une vraie mise à jour ;
4. avancer délibérément `reviewed_commit` et le lien correspondant dans l'article
   après cette revue, jamais en réponse automatique à une nouvelle version ;
5. passer les tests du site, le lint/build et les contrôles de la veille, puis
   fusionner par PR et fermer l'issue manuellement.

Une variation du code n'est pas une preuve que l'article est faux. Une CI verte
n'est pas une validation sur carte. La notification sert à empêcher l'oubli de
la revue, pas à fabriquer des affirmations ou publier des métriques en direct.

Le mode interne `--publish` est réservé à Actions dans le dépôt recevant les
issues ; un token personnel lancé en local n'est pas un mode de publication
supporté. Utiliser le dispatch du workflow garde la même identité de bot.
