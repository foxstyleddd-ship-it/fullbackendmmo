# Guide d'Intégration avec Unreal Engine

## 🚀 Démarrage Rapide

### 1. Démarrer l'Infrastructure (Docker)

```bash
# Démarrer PostgreSQL et Redis
docker-compose up -d postgres redis

# Vérifier que tout fonctionne
docker-compose ps
```

**Vérification :**
- ✅ `hpmmo-postgres` : Status `Up (healthy)`
- ✅ `hpmmo-redis` : Status `Up (healthy)`

### 2. Exécuter les Migrations

```bash
# Installer l'outil de migration (une seule fois)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Exécuter les migrations
migrate -path migrations -database "postgres://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" up

# Vérifier la version actuelle
migrate -path migrations -database "postgres://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" version
```

**Résultat attendu :**
```
2 (002_spells.up.sql)
```

### 3. Créer un Compte de Test

```bash
# Lancer l'API temporairement
go run cmd/api/main.go &

# Créer un compte
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@hogwarts.com",
    "username": "testplayer",
    "password": "Password123!"
  }'

# Créer un personnage
# 1. Login pour obtenir le token
TOKEN=$(curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@hogwarts.com", "password": "Password123!"}' \
  | jq -r '.data.access_token')

# 2. Créer le personnage
curl -X POST http://localhost:8080/v1/characters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Harry Potter",
    "house": "gryffindor",
    "appearance": {}
  }'
```

### 4. Démarrer les Serveurs

**Terminal 1 - API Server :**
```bash
go run cmd/api/main.go
```

**Terminal 2 - Zone Server (WebSocket) :**
```bash
go run cmd/zone/main.go
```

**Terminal 3 - Admin Panel (Optionnel) :**
```bash
cd admin-panel
npm run dev
```

## 🎮 Intégration Unreal Engine

### Architecture de Communication

```
Unreal Engine 5
    │
    ├── HTTP REST API (Port 8080)
    │   ├── Login/Register
    │   ├── Character Management
    │   ├── GET Spells
    │   ├── GET Inventory
    │   └── GET Character Stats
    │
    └── WebSocket (Port 9000)
        ├── Real-time notifications
        ├── GM commands (teleport, grant items/spells)
        └── Zone coordination
```

### Étape 1 : Login (HTTP)

**Blueprint Unreal :**

```
Event BeginPlay
  │
  ├── HTTP Request: POST http://localhost:8080/v1/auth/login
  │   Body: {"email": "test@hogwarts.com", "password": "Password123!"}
  │
  ├── On Response Received
  │   ├── Parse JSON
  │   ├── Save: AccessToken = response.data.access_token
  │   ├── Save: AccountID = response.data.account.id
  │   └── Save: Characters = response.data.characters
  │
  └── Call: SelectCharacter(Characters[0].id)
```

**Exemple de réponse :**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "account": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "test@hogwarts.com",
      "username": "testplayer",
      "role": "player"
    },
    "characters": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Harry Potter",
        "house": "gryffindor",
        "grade": 1,
        "level": 1,
        "zone_id": "hogwarts_main"
      }
    ]
  }
}
```

### Étape 2 : Sélectionner un Personnage (HTTP)

```
SelectCharacter(CharacterID)
  │
  ├── HTTP Request: POST http://localhost:8080/v1/characters/{CharacterID}/select
  │   Header: Authorization: Bearer {AccessToken}
  │
  ├── On Response Received
  │   ├── Update: AccessToken = response.data.access_token (nouveau token avec character_id)
  │   ├── Save: CharacterID = response.data.character_id
  │   └── Save: Permissions = response.data.permissions
  │
  └── Call: LoadCharacterData(CharacterID)
```

### Étape 3 : Charger les Données du Personnage (HTTP)

**A. Récupérer les Détails :**
```
GET http://localhost:8080/v1/characters/{CharacterID}
Header: Authorization: Bearer {AccessToken}
```

**B. Récupérer les Spells :**
```
GET http://localhost:8080/v1/characters/{CharacterID}/spells
Header: Authorization: Bearer {AccessToken}
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "character_id": "550e8400-e29b-41d4-a716-446655440000",
    "spells": [
      {
        "id": "abc123...",
        "spell_id": "lumos",
        "name": "Lumos",
        "description": "Crée une source de lumière",
        "spell_school": "charms",
        "mana_cost": 10,
        "cooldown_seconds": 2,
        "is_forbidden": false,
        "proficiency_level": 1,
        "times_cast": 0,
        "required_grade": 1
      },
      {
        "id": "def456...",
        "spell_id": "expelliarmus",
        "name": "Expelliarmus",
        "description": "Désarme l'adversaire",
        "spell_school": "defense_against_dark_arts",
        "mana_cost": 20,
        "cooldown_seconds": 5,
        "is_forbidden": false,
        "proficiency_level": 2,
        "times_cast": 15,
        "required_grade": 1
      }
    ],
    "count": 2
  }
}
```

**C. Récupérer l'Inventaire :**
```
GET http://localhost:8080/v1/characters/{CharacterID}/inventory
Header: Authorization: Bearer {AccessToken}
```

**Dans Unreal :**
```
LoadCharacterData(CharacterID)
  │
  ├── Parallel Requests:
  │   ├── GET /characters/{ID}        → Character Stats
  │   ├── GET /characters/{ID}/spells → Spell List
  │   └── GET /characters/{ID}/inventory → Items
  │
  ├── Parse All Responses
  │
  ├── For Each Spell:
  │   └── Create Spell Ability (ACF framework)
  │
  ├── For Each Item:
  │   └── Add to Inventory Widget
  │
  └── Call: ConnectToZoneServer()
```

### Étape 4 : Connexion WebSocket au Zone Server

**Blueprint WebSocket :**

```
ConnectToZoneServer()
  │
  ├── WebSocket Connect: ws://localhost:9000
  │
  ├── On Connected:
  │   └── Send Authentication:
  │       {
  │         "type": "authenticate",
  │         "token": "{AccessToken}"
  │       }
  │
  ├── On Message Received:
  │   └── Parse JSON → Route by "type"
  │
  └── On Disconnected:
      └── Show Reconnection Dialog
```

**Messages WebSocket Entrants (Server → Client) :**

```json
// 1. Spawn Initial
{
  "type": "player_spawn",
  "character": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Harry Potter",
    "level": 1,
    "grade": 1
  },
  "position": {"x": 0, "y": 0, "z": 100},
  "zone_id": "hogwarts_main"
}

// 2. Spell Accordé (via Admin Panel)
{
  "type": "spell_granted",
  "spell_id": "expecto_patronum",
  "spell_data": {
    "name": "Expecto Patronum",
    "spell_school": "defense_against_dark_arts",
    "mana_cost": 100,
    "cooldown_seconds": 60,
    "required_grade": 5
  }
}

// 3. Item Accordé
{
  "type": "item_granted",
  "item_def_id": "elder_wand",
  "quantity": 1
}

// 4. Téléportation
{
  "type": "teleport",
  "zone_id": "hogsmeade",
  "position": {"x": 1000, "y": 2000, "z": 50}
}
```

**Handler dans Unreal :**
```
OnWebSocketMessage(JsonString)
  │
  ├── Parse JSON
  │
  └── Switch on Type:
      ├── "player_spawn" → SpawnPlayer(position)
      ├── "spell_granted" → AddSpellToPlayer(spell_data)
      ├── "item_granted" → AddItemToInventory(item_def_id, quantity)
      ├── "teleport" → TeleportToZone(zone_id, position)
      └── default → Log Warning
```

### Étape 5 : Cast d'un Spell (ACF Framework)

**Exemple avec ACF (Action Combat Framework) :**

```cpp
// Dans votre Character class

void AMyCharacter::CastSpell(FString SpellID)
{
    // 1. Trouver le spell dans la liste
    FSpellData* Spell = LearnedSpells.FindByPredicate([&](const FSpellData& S) {
        return S.SpellID == SpellID;
    });

    if (!Spell) {
        UE_LOG(LogTemp, Warning, TEXT("Spell not found: %s"), *SpellID);
        return;
    }

    // 2. Vérifier les prérequis (local)
    if (CurrentMana < Spell->ManaCost) {
        ShowError("Pas assez de mana!");
        return;
    }

    if (IsSpellOnCooldown(SpellID)) {
        ShowError("Sort en cooldown!");
        return;
    }

    // 3. Consommer les ressources (local)
    CurrentMana -= Spell->ManaCost;
    StartCooldown(SpellID, Spell->CooldownSeconds);

    // 4. Exécuter l'animation et VFX (ACF)
    UACFAbilityComponent* AbilityComp = GetComponentByClass<UACFAbilityComponent>();
    if (AbilityComp) {
        // ACF gère l'animation, le projectile, les dégâts, etc.
        AbilityComp->TryActivateAbilityByTag(FGameplayTag::RequestGameplayTag(FName(*SpellID)));
    }

    // 5. (Optionnel) Notifier le backend pour tracking
    if (WebSocket && WebSocket->IsConnected()) {
        TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
        JsonObject->SetStringField("type", "spell_cast");
        JsonObject->SetStringField("spell_id", SpellID);

        FString OutputString;
        TSharedRef<TJsonWriter<>> Writer = TJsonWriterFactory<>::Create(&OutputString);
        FJsonSerializer::Serialize(JsonObject.ToSharedRef(), Writer);

        WebSocket->Send(OutputString);
    }
}
```

**Structure de données Unreal :**

```cpp
// Dans votre GameInstance ou PlayerState

USTRUCT(BlueprintType)
struct FSpellData
{
    GENERATED_BODY()

    UPROPERTY(BlueprintReadWrite)
    FString SpellID;

    UPROPERTY(BlueprintReadWrite)
    FString Name;

    UPROPERTY(BlueprintReadWrite)
    FString Description;

    UPROPERTY(BlueprintReadWrite)
    FString SpellSchool;

    UPROPERTY(BlueprintReadWrite)
    int32 ManaCost;

    UPROPERTY(BlueprintReadWrite)
    int32 CooldownSeconds;

    UPROPERTY(BlueprintReadWrite)
    int32 ProficiencyLevel;

    UPROPERTY(BlueprintReadWrite)
    int32 TimesCast;

    UPROPERTY(BlueprintReadWrite)
    bool bIsForbidden;
};

UCLASS()
class UMyGameInstance : public UGameInstance
{
    GENERATED_BODY()

public:
    UPROPERTY(BlueprintReadWrite)
    FString AccessToken;

    UPROPERTY(BlueprintReadWrite)
    FString CharacterID;

    UPROPERTY(BlueprintReadWrite)
    TArray<FSpellData> LearnedSpells;

    UPROPERTY(BlueprintReadWrite)
    TArray<FString> Permissions;
};
```

## 🧪 Scénario de Test Complet

### 1. Test du Login

```bash
# Dans un terminal
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@hogwarts.com", "password": "Password123!"}' \
  | jq
```

**Dans Unreal :**
- Créer un Widget de login
- Tester la connexion
- Vérifier que le token est bien stocké

### 2. Test des Spells

**Depuis l'Admin Panel (http://localhost:5173) :**
1. Login avec compte admin
2. Chercher "Harry Potter"
3. Accorder le spell "expecto_patronum"

**Dans Unreal :**
1. Le WebSocket reçoit `{"type": "spell_granted", ...}`
2. Ajouter le spell à la liste
3. Afficher dans l'UI
4. Tester le cast

### 3. Test de Téléportation

**Admin Panel :**
1. Sélectionner zone "hogsmeade"
2. Cliquer "Teleport"

**Unreal :**
1. Reçoit `{"type": "teleport", "zone_id": "hogsmeade", ...}`
2. Stream level "Hogsmeade"
3. Placer le joueur à la position indiquée

## 📊 Logs et Debugging

### Vérifier les Logs Backend

```bash
# Logs API
docker-compose logs -f api

# Logs Zone Server
docker-compose logs -f zone

# Logs PostgreSQL
docker-compose logs -f postgres
```

### Tester les Endpoints Manuellement

```bash
# Liste des spells disponibles
curl http://localhost:8080/v1/spells | jq

# Spells d'un personnage (avec token)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/v1/characters/YOUR_CHARACTER_ID/spells | jq
```

### WebSocket Testing Tool

Utiliser [Postman](https://www.postman.com/) ou [wscat](https://www.npmjs.com/package/wscat) :

```bash
# Installer wscat
npm install -g wscat

# Se connecter au Zone Server
wscat -c ws://localhost:9000

# Envoyer l'authentification
> {"type": "authenticate", "token": "YOUR_TOKEN_HERE"}

# Recevoir les messages
< {"type": "player_spawn", ...}
```

## 🎯 Prochaines Étapes

1. **Créer un compte de test** avec l'admin panel
2. **Grant des spells** au personnage
3. **Implémenter le login** dans Unreal (Blueprint ou C++)
4. **Charger les spells** depuis l'API
5. **Connecter le WebSocket** pour les notifications temps réel
6. **Intégrer ACF** pour le système de combat avec les spells

## ❓ FAQ

**Q: Le backend valide-t-il les dégâts ?**
R: Non, Unreal gère tout le combat. Le backend stocke juste les spells appris et les stats.

**Q: Comment gérer le multijoueur ?**
R: Unreal Replication pour la synchro entre joueurs. Backend = persistance uniquement.

**Q: Dois-je notifier le backend à chaque cast ?**
R: C'est optionnel. Utile pour tracking/analytics, mais pas obligatoire pour le gameplay.

**Q: Les cooldowns sont gérés où ?**
R: Localement dans Unreal pour le gameplay. Backend ne les valide pas.
