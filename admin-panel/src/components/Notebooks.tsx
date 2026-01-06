import { useState } from 'react';
import { apiService } from '../services/api';

interface Notebook {
  id: string;
  character_id: string;
  title: string;
  content: string;
  subject: string | null;
  created_at: string;
  updated_at: string;
}

export default function Notebooks() {
  const [characterId, setCharacterId] = useState('');
  const [notebooks, setNotebooks] = useState<Notebook[]>([]);
  const [selectedNotebook, setSelectedNotebook] = useState<Notebook | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Form fields
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [subject, setSubject] = useState('');

  const subjects = [
    { value: '', label: 'Aucune matière' },
    { value: 'charms', label: 'Enchantements' },
    { value: 'transfiguration', label: 'Métamorphose' },
    { value: 'potions', label: 'Potions' },
    { value: 'defense_against_dark_arts', label: 'Défense contre les Forces du Mal' },
    { value: 'herbology', label: 'Botanique' },
    { value: 'astronomy', label: 'Astronomie' },
    { value: 'history_of_magic', label: 'Histoire de la Magie' },
    { value: 'care_of_magical_creatures', label: 'Soins aux Créatures Magiques' },
    { value: 'divination', label: 'Divination' },
    { value: 'ancient_runes', label: 'Runes Anciennes' },
    { value: 'arithmancy', label: 'Arithmancie' },
    { value: 'muggle_studies', label: 'Étude des Moldus' },
    { value: 'flying', label: 'Vol' },
  ];

  const loadNotebooks = async () => {
    if (!characterId) {
      setError('Veuillez entrer un ID de personnage');
      return;
    }

    try {
      setLoading(true);
      setError('');
      setMessage(null);
      const data = await apiService.getCharacterNotebooks(characterId);
      setNotebooks(data.notebooks || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Échec du chargement des carnets');
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setTitle('');
    setContent('');
    setSubject('');
    setSelectedNotebook(null);
    setIsCreating(false);
    setIsEditing(false);
  };

  const handleCreate = () => {
    setIsCreating(true);
    setIsEditing(false);
    resetForm();
  };

  const handleEdit = (notebook: Notebook) => {
    setSelectedNotebook(notebook);
    setTitle(notebook.title);
    setContent(notebook.content);
    setSubject(notebook.subject || '');
    setIsEditing(true);
    setIsCreating(false);
  };

  const handleSave = async () => {
    if (!characterId) {
      setMessage({ type: 'error', text: 'ID de personnage requis' });
      return;
    }

    if (!title.trim()) {
      setMessage({ type: 'error', text: 'Le titre est requis' });
      return;
    }

    try {
      setLoading(true);
      setMessage(null);

      if (isEditing && selectedNotebook) {
        // Update existing notebook
        await apiService.updateNotebook(selectedNotebook.id, {
          title: title.trim(),
          content: content.trim(),
          subject: subject || undefined,
        });
        setMessage({ type: 'success', text: 'Carnet mis à jour avec succès' });
      } else {
        // Create new notebook
        await apiService.createNotebook({
          character_id: characterId,
          title: title.trim(),
          content: content.trim(),
          subject: subject || undefined,
        });
        setMessage({ type: 'success', text: 'Carnet créé avec succès' });
      }

      resetForm();
      loadNotebooks();
    } catch (err: any) {
      setMessage({
        type: 'error',
        text: err.response?.data?.error?.message || 'Échec de la sauvegarde du carnet',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (notebookId: string) => {
    if (!confirm('Êtes-vous sûr de vouloir supprimer ce carnet ?')) return;

    try {
      setLoading(true);
      setMessage(null);
      await apiService.deleteNotebook(notebookId);
      setMessage({ type: 'success', text: 'Carnet supprimé' });
      resetForm();
      loadNotebooks();
    } catch (err: any) {
      setMessage({
        type: 'error',
        text: err.response?.data?.error?.message || 'Échec de la suppression',
      });
    } finally {
      setLoading(false);
    }
  };

  const getSubjectLabel = (value: string | null) => {
    if (!value) return 'Aucune matière';
    return subjects.find(s => s.value === value)?.label || value;
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">📖 Carnets d'École</h1>

      {/* Character ID Input */}
      <div className="mb-6 bg-white p-4 rounded-lg shadow">
        <div className="flex gap-2">
          <input
            type="text"
            value={characterId}
            onChange={(e) => setCharacterId(e.target.value)}
            placeholder="ID du personnage (UUID)"
            className="flex-1 px-3 py-2 border border-gray-300 rounded"
          />
          <button
            onClick={loadNotebooks}
            disabled={loading}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            Charger
          </button>
          <button
            onClick={handleCreate}
            disabled={loading || !characterId}
            className="px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:bg-gray-400"
          >
            + Nouveau Carnet
          </button>
        </div>
      </div>

      {/* Messages */}
      {message && (
        <div
          className={`mb-4 p-3 rounded border ${
            message.type === 'success'
              ? 'bg-green-100 border-green-400 text-green-700'
              : 'bg-red-100 border-red-400 text-red-700'
          }`}
        >
          {message.text}
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {/* Create/Edit Form */}
      {(isCreating || isEditing) && (
        <div className="mb-6 bg-white p-6 rounded-lg shadow">
          <h2 className="text-2xl font-bold mb-4">
            {isEditing ? 'Modifier le carnet' : 'Nouveau carnet'}
          </h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Titre</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Titre du carnet..."
                className="w-full px-3 py-2 border border-gray-300 rounded"
                maxLength={128}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Matière (optionnel)</label>
              <select
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded"
              >
                {subjects.map(s => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Contenu</label>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Prenez vos notes ici..."
                rows={10}
                className="w-full px-3 py-2 border border-gray-300 rounded font-mono text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleSave}
                disabled={loading}
                className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
              >
                Enregistrer
              </button>
              <button
                onClick={resetForm}
                disabled={loading}
                className="px-6 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:bg-gray-400"
              >
                Annuler
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Notebooks List */}
      {!isCreating && !isEditing && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {notebooks.map((notebook) => (
            <div key={notebook.id} className="bg-white p-4 rounded-lg shadow hover:shadow-md transition">
              <div className="flex items-start justify-between mb-2">
                <h3 className="text-lg font-semibold text-gray-900">{notebook.title}</h3>
                {notebook.subject && (
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                    {getSubjectLabel(notebook.subject)}
                  </span>
                )}
              </div>
              <p className="text-sm text-gray-600 mb-3 line-clamp-3">
                {notebook.content || <em>Carnet vide</em>}
              </p>
              <div className="text-xs text-gray-500 mb-3">
                Modifié: {new Date(notebook.updated_at).toLocaleString()}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => handleEdit(notebook)}
                  className="flex-1 px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                >
                  Modifier
                </button>
                <button
                  onClick={() => handleDelete(notebook.id)}
                  className="px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700"
                >
                  Supprimer
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {!isCreating && !isEditing && notebooks.length === 0 && characterId && !loading && (
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <p className="text-gray-500 mb-4">Aucun carnet trouvé pour ce personnage</p>
          <button
            onClick={handleCreate}
            className="px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Créer le premier carnet
          </button>
        </div>
      )}
    </div>
  );
}
