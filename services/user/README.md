# User Service

Service NestJS responsable de l'authentification et de la gestion des utilisateurs.

## Lancer le service

```bash
npm install
npm run start:dev
```

---

## Tests

### Lancer les tests

```bash
# Tous les tests unitaires
npm run test

# Avec le rapport de couverture
npm run test:cov

# Mode watch (relance à chaque modification)
npm run test:watch
```

### Résultats attendus

```
Tests:   21 passed
auth.service.ts   → 98% de couverture
user.service.ts   → 100% de couverture
```

---

## Comprendre les tests unitaires

### C'est quoi un test unitaire ?

Un test unitaire vérifie **un seul service isolé**, sans base de données, sans réseau.
On remplace toutes les dépendances par des **mocks** (faux objets qu'on contrôle).

```
Sans test unitaire :        Avec test unitaire :
AuthService                 AuthService
    ↓                           ↓
PostgreSQL (vraie DB)       mockUserRepo (faux objet)
    ↓                           ↓
réseau, lenteur, état       réponse qu'on décide nous-mêmes
```

### Structure d'un test

Chaque test suit 3 étapes :

```typescript
it('should throw ConflictException if email already exists', async () => {
  // 1. Arrange — on prépare le mock
  mockUserRepo.findOne.mockResolvedValue(mockUser); // simule "email déjà en base"

  // 2. Act + Assert — on appelle le service et on vérifie
  await expect(
    service.register({ username: 'test', email: 'test@example.com', password: '123' }),
  ).rejects.toThrow(ConflictException);
});
```

---

## Ce qui est testé

### `auth.service.spec.ts`

| Méthode | Cas testés |
|---|---|
| `register` | Email déjà existant → `ConflictException` |
| `register` | Email libre → crée l'utilisateur et retourne les tokens |
| `login` | Utilisateur introuvable → `UnauthorizedException` |
| `login` | Mauvais mot de passe → `UnauthorizedException` |
| `login` | Credentials valides → retourne les tokens |
| `validateToken` | Token invalide → `{ valid: false }` |
| `validateToken` | Token valide mais utilisateur supprimé → `{ valid: false }` |
| `validateToken` | Token valide → `{ valid: true, user }` |
| `logout` | Révoque tous les refresh tokens de l'utilisateur |
| `refresh` | Token introuvable → `UnauthorizedException` |
| `refresh` | Token expiré → `UnauthorizedException` |
| `refresh` | Token valide → révoque l'ancien et retourne de nouveaux tokens |

### `user.service.spec.ts`

| Méthode | Cas testés |
|---|---|
| `findById` | Utilisateur introuvable → `NotFoundException` |
| `findById` | Utilisateur trouvé → retourne les données sans `password_hash` |
| `update` | Requester ≠ propriétaire → `ForbiddenException` |
| `update` | Utilisateur introuvable → `NotFoundException` |
| `update` | Mise à jour du `username` |
| `update` | Mise à jour de l'`avatar_url` |
| `update` | Les champs non fournis ne sont pas modifiés |
| `update` | Le résultat ne contient jamais `password_hash` |

---

## Concepts clés

### `jest.fn()` — fonction mock

Remplace une vraie fonction par une fonction vide qu'on contrôle :

```typescript
const mockUserRepo = {
  findOne: jest.fn(), // ne fait rien par défaut
};
```

### `mockResolvedValue` vs `mockReturnValue`

```typescript
mockUserRepo.findOne.mockResolvedValue(user); // pour les fonctions async (retourne une Promise)
mockUserRepo.create.mockReturnValue(user);    // pour les fonctions sync
```

### `jest.clearAllMocks()` dans `beforeEach`

Remet tous les mocks à zéro avant chaque test pour éviter que les appels d'un test contaminent le suivant.

### `rejects.toThrow` — tester une exception

```typescript
await expect(service.login({ ... })).rejects.toThrow(UnauthorizedException);
```

### `jest.mock('bcrypt')` — mocker un module entier

Remplace toutes les fonctions de bcrypt par des `jest.fn()`.
Nécessaire car bcrypt est lent (vraies opérations crypto).
On contrôle ensuite ce que `bcrypt.hash` et `bcrypt.compare` retournent dans chaque test.

---

## Ce qui n'est pas encore testé

| Type | Quand |
|---|---|
| Tests d'intégration (vraie DB) | Quand l'infra Azure est disponible |
| Tests E2E (API complète) | Quand tous les services tournent ensemble |
| Controllers | Couvert par les tests E2E |
