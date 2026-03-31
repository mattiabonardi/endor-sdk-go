# Aggregation Pipeline — Albero delle Dipendenze

## Struttura root della pipeline

```
AggregationPipeline  []json.RawMessage
│
├── EntityPipelineStage          (uno o più — identificati dalla chiave "entity")
│   ├── entity   string          → nome dell'entità nel RepositoryRegistry
│   └── pipeline []StageSpec     → sequenza di stage applicati ai documenti
│       │
│       ├── $match  ─────────────────────────────────────────────────────────┐
│       │   │  (primo $match: push-down al repository come filtro)           │
│       │   │  (ulteriori $match: eseguiti in-memory)                        │
│       │   │                                                                │
│       │   ├── <campo>: <valore>          → uguaglianza diretta             │
│       │   ├── <campo>: { $eq: v }        → uguale a v                      │
│       │   ├── <campo>: { $ne: v }        → diverso da v                    │
│       │   ├── <campo>: { $gt: v }        → maggiore di v                   │
│       │   ├── <campo>: { $gte: v }       → maggiore o uguale a v           │
│       │   ├── <campo>: { $lt: v }        → minore di v                     │
│       │   ├── <campo>: { $lte: v }       → minore o uguale a v             │
│       │   ├── <campo>: { $in: [...] }    → valore presente nell'array      │
│       │   ├── <campo>: { $nin: [...] }   → valore assente nell'array       │
│       │   ├── <campo>: { $exists: bool } → presenza/assenza del campo      │
│       │   ├── <campo>: { $regex: str }   → sottostringa (strings.Contains) │
│       │   │                                                                │
│       │   ├── $and: [ {clause}, ... ]    → tutte le clausole true          │
│       │   ├── $or:  [ {clause}, ... ]    → almeno una clausola true        │
│       │   └── $nor: [ {clause}, ... ]    → nessuna clausola true           │
│       │                                                                    │
│       └── $group                                                           │
│           ├── id: <expr>                 → chiave di raggruppamento        │
│           │     └── "$campo"  →  valore del campo (field ref)              │
│           │         <literal> →  valore costante                           │
│           │                                                                │
│           └── <outputField>: { <accumulator>: <expr> }                    │
│               │                                                            │
│               ├── $sum:      "$campo" | 1    → somma numerica              │
│               ├── $avg:      "$campo"        → media numerica              │
│               ├── $min:      "$campo"        → valore minimo               │
│               ├── $max:      "$campo"        → valore massimo              │
│               ├── $first:    "$campo"        → valore del primo documento  │
│               ├── $last:     "$campo"        → valore dell'ultimo doc.     │
│               ├── $push:     "$campo"        → array di tutti i valori     │
│               ├── $addToSet: "$campo"        → array deduplicated          │
│               └── $count:    (qualsiasi)     → numero di documenti          │
│                                                                            │
└── Operatori top-level          (vengono dopo le EntityPipelineStage)       │
    │                                                                        │
    └── $mergeResults                                                        │
        ├── on:     string        → campo di join tra i risultati entità     │
        └── fields: []string      → campi da copiare (vuoto = tutti)         │
```

---

## Regole di parsing e ordine di esecuzione

```
AggregationEngine.Execute
│
├── 1. Per ogni elemento del pipeline (left-to-right):
│   │
│   ├── Se ha la chiave "entity"  →  EntityPipelineStage
│   │     ├── Recupera repository dal RegistryRegistry
│   │     ├── Primo stage è $match?  →  push-down a ListDocuments (filtro)
│   │     └── Stadi rimanenti  →  applicati in-memory nell'ordine dichiarato
│   │           ├── $match  →  applyMatch (filtra documenti)
│   │           └── $group  →  applyGroup (raggruppa e accumula)
│   │
│   └── Altrimenti  →  operatore top-level (mappa chiave→valore)
│         └── $mergeResults  →  mergeResults(entityResults, entityOrder, opts)
│                                 ├── Itera entità nell'ordine di apparizione
│                                 ├── Usa "on" come chiave di join
│                                 └── Copia solo "fields" (o tutti se vuoto)
│
└── 2. Risultato finale:
      ├── $mergeResults presente  →  restituisce il risultato del merge
      ├── Una sola entità         →  restituisce il suo result set direttamente
      └── Altrimenti              →  restituisce slice vuota
```

---

## Esempi d'uso

### 1. Filtraggio semplice
```json
[
  {
    "entity": "order",
    "pipeline": [
      { "$match": { "status": "completed" } }
    ]
  }
]
```

### 2. Filtro composto con operatori logici
```json
[
  {
    "entity": "order",
    "pipeline": [
      {
        "$match": {
          "$and": [
            { "status": "completed" },
            { "amount": { "$gte": 100 } }
          ]
        }
      }
    ]
  }
]
```

### 3. Raggruppamento con accumulatori
```json
[
  {
    "entity": "order",
    "pipeline": [
      { "$match": { "status": "completed" } },
      {
        "$group": {
          "id": "$customerId",
          "totalSpent": { "$sum": "$amount" },
          "orderCount": { "$count": 1 },
          "avgOrder":   { "$avg": "$amount" }
        }
      }
    ]
  }
]
```

### 4. Join tra due entità con `$mergeResults`
```json
[
  {
    "entity": "order",
    "pipeline": [
      { "$group": { "id": "$customerId", "total": { "$sum": "$amount" } } }
    ]
  },
  {
    "entity": "customer",
    "pipeline": [
      { "$match": { "active": true } }
    ]
  },
  {
    "$mergeResults": {
      "on": "id",
      "fields": ["total", "name", "country"]
    }
  }
]
```

---

## Risoluzione delle espressioni (`resolveExpr`)

| Valore `expr`   | Risultato                                      |
|-----------------|------------------------------------------------|
| `"$campo"`      | valore di `doc["campo"]` (field reference)     |
| `"$a.b.c"`      | valore di `doc["a"]["b"]["c"]` (dot-notation)  |
| qualsiasi altro | il valore stesso come costante letterale        |

---

## Vincoli e comportamenti da tenere a mente

| Situazione | Comportamento |
|---|---|
| Entità non registrata nel RepositoryRegistry | Errore: `entity "X" not found in repository registry` |
| `$regex` | Usa `strings.Contains` — **non** è una regex reale |
| `$mergeResults` con `fields` vuoto | Tutti i campi vengono copiati; conflitti: l'entità più recente vince |
| `$mergeResults` con entità mancante | Lo stage viene silenziosamente saltato (nessun errore) |
| Pipeline senza `$mergeResults` e più di una entità | Viene restituita una slice vuota |
| Primo `$match` in un entity stage | Push-down al repository — deve essere un oggetto semplice |
