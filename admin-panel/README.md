# 🎮 Harry Potter MMO - Admin Panel

Panel d'administration web pour gérer le backend MMO Harry Potter.

## 🚀 Quick Start

### Installation

```bash
cd admin-panel
npm install
```

### Configuration

Créer un fichier `.env`:

```bash
cp .env.example .env
```

### Développement

```bash
npm run dev
```

Accessible sur `http://localhost:5173`

### Build Production

```bash
npm run build
```

## 🔐 Authentification

Nécessite un rôle: `gm`, `admin`, ou `superadmin`

Créer un admin:
```sql
UPDATE accounts SET role = 'admin' WHERE email = 'your@email.com';
```

## ✨ Fonctionnalités

- **Recherche & Gestion Personnages**
- **Modification de Grade** (1-100)
- **Téléportation** (7 zones disponibles)
- **Gestion d'Inventaire** (voir + donner items)
- **Audit Logs** (historique complet)
