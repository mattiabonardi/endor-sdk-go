# Aggregation Pipeline Distribuita e Query Layer Multi-Microservizi

## 1. Concetto

L’idea è di avere un **query layer centralizzato** sopra un’architettura a microservizi, dove:

* Ogni microservizio gestisce **entità differenti** e sa come aggregare i propri dati.
* Il **query layer** non conosce dove risiedono i dati e non accede direttamente ai database dei microservizi.
* Il query layer costruisce una **pipeline distribuita** e invia ogni sottopipeline al microservizio corrispondente.
* I microservizi ritornano solo il **risultato aggregato**, e il query layer combina i risultati (merge, union, ordinamento) lato applicazione.

Vantaggi:

* Evita join costosi tra microservizi.
* Scalabile e parallelo.
* Astrae il dettaglio dei dati e della loro ubicazione.

---

## 2. Sintassi “Mongo-like” distribuita

La pipeline distribuita mantiene uno stile simile alla Aggregation Pipeline di MongoDB, con operatori estesi per più microservizi:

```js
[
  {
    entity: "order",
    pipeline: [
      { $match: { status: "completed" } },
      { $group: { _id: "$customerId", totalSpent: { $sum: "$amount" } } }
    ]
  },
  {
    entity: "review",
    pipeline: [
      { $match: { rating: { $gte: 4 } } },
      { $group: { _id: "$userId", positiveReviews: { $sum: 1 } } }
    ]
  },
  {
    $mergeResults: {
      on: "_id",                   // chiave logica per combinare i documenti
      fields: ["totalSpent", "positiveReviews"]
    }
  },
  { $sort: { totalSpent: -1 } }
]
```

### 2.1 Descrizione operatori

| Operatore            | Funzione                                                               |
| -------------------- | ---------------------------------------------------------------------- |
| `entity`             | Specifica il microservizio / entità a cui inviare la sottopipeline     |
| `pipeline`           | Lista di stage Mongo-like (`$match`, `$group`, `$project`, etc.)       |
| `$mergeResults`      | Combina i risultati aggregati di più microservizi su una chiave comune |
| `$unionResults`      | Unisce risultati di microservizi indipendenti senza correlazione       |
| `$compute`           | Calcola campi basati sui dati combinati di microservizi diversi        |
| `$sort` / `$project` | Operazioni post-merge sul risultato finale                             |

---

## 3. Differenze rispetto a MongoDB `$lookup`

| Aspetto                 | `$lookup` (MongoDB)          | `$mergeResults` (Query Layer)           |
| ----------------------- | ---------------------------- | --------------------------------------- |
| Dove avviene            | Interno al DB                | Lato query layer sopra microservizi     |
| Input                   | Dati grezzi                  | Risultati aggregati dai microservizi    |
| Output                  | Documenti con array di match | Documenti combinati, campi fusi         |
| Necessità di join reale | Sì                           | No, merge logico lato layer             |
| Scalabilità             | Limitata al singolo DB       | Scalabile su più microservizi paralleli |

💡 Regola pratica: `$mergeResults` può emulare un `$lookup` quando si combinano dati correlati da microservizi diversi.

---

## 4. Esempio concettuale di pipeline distribuita

```text
Query Layer
 ├─ Pipeline Orders Service
 │    └─ $match / $group / $project
 ├─ Pipeline Reviews Service
 │    └─ $match / $group
 ├─ $mergeResults on _id
 └─ $sort / $project finale
```

* Ogni microservizio calcola la propria aggregazione
* Il query layer combina e trasforma i risultati aggregati
* Il client riceve un dataset unificato senza conoscere dove risiedono i dati