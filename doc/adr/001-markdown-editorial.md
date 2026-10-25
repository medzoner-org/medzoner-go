# Catalogue éditorial Markdown dans le monolithe Go

## Contexte

Les articles étaient chacun un template HTML, une route et un handler spécifiques.
Le sitemap statique et les métadonnées pouvaient diverger du contenu. Le besoin
est un backend de média tech édité dans Git, pas un CMS ni une refonte graphique.

## Décision

Conserver le monolithe hexagonal existant et ajouter un contexte éditorial :

```
HTTP articles / liste / RSS / sitemap
  → application/query.Editorial (publication, filtres, pagination)
    → ports/editorial.Repository
      ← adapters/editorial.Repository (YAML + Markdown embarqués)
```

Le modèle `domains/editorial` ne dépend ni de HTTP, ni de SQL, ni du moteur
Markdown. `resources` embarque les sources au build ; le provider whyor charge et
valide le catalogue une fois au démarrage. Les handlers ne construisent aucun
chemin de fichier à partir des requêtes.

Goldmark transforme le Markdown à la frontière de présentation. Le HTML brut et
les protocoles de liens dangereux restent désactivés. Les attributs issus du
Markdown sont supprimés, sauf les identifiants d'ancres validés ; les classes
d'encadrés et boutons sont choisies par le renderer, jamais par le contenu.
Le renderer produit aussi les entrées du sommaire depuis les titres du document.
Seul son résultat est
marqué `template.HTML`. Le JSON-LD passe par `encoding/json` et les flux par
`encoding/xml`, sans concaténation de métadonnées dans du code HTML/XML.

Une requête commune `Published` exclut les brouillons et les publications futures
avant toute lecture publique : détail, liste, sitemap et RSS. Les URL historiques
restent stables. Un slug inconnu, brouillon ou futur donne le même statut 404.

## Alternatives et conséquences

- Un CMS/base éditoriale demanderait authentification, administration, migrations
  et sauvegardes sans besoin avéré. Git fournit déjà revue et historique.
- Servir des fichiers éditables sur disque introduirait des différences entre les
  conteneurs. Les sources embarquées rendent une livraison reproductible, au prix
  d'un rebuild pour publier ou modifier du contenu.
- Pas de commandes applicatives d'écriture ni d'event bus éditorial : les commits
  Git constituent le workflow d'écriture ; les use cases Go sont des queries.
- La une et les liens de sélection du header restent des sélections manuelles :
  ce changement est backend. `/articles` est le catalogue exhaustif automatique.
- Le RSS contient les 20 publications les plus récentes. Le catalogue est chargé
  en mémoire et les listes sont paginées ; à forte volumétrie, un autre adapter
  de repository pourra remplacer les fichiers sans changer les règles métier.
- Les brouillons ne sont pas publics via HTTP mais sont présents dans le dépôt et
  le binaire : ne jamais y stocker de secrets.

## Migration et validation

Les deux corps HTML sont convertis en Markdown ; les blocs de code, liens, images
et identifiants de section sont conservés. Les dates proviennent de l'historique
Git des articles. Un template unique remplace les templates par article.
Le sitemap statique est remplacé par `/sitemap.xml` ; RSS est publié à `/feed.xml`.
La CI ne doit pas ignorer les Markdown éditoriaux ni leurs images : une modification
de ces sources déclenche désormais le build comme une modification de code.
Les tests couvrent la validation du catalogue, la non-publication, le rendu sûr,
les URLs historiques, les filtres, les extrêmes de pagination et les flux XML.
