import { useState } from 'react';
import { apiService } from '../services/api';
import type { CharacterDetail, InventoryItem, CharacterSpell } from '../types';

export default function CharacterDetail() {
  const [characterId, setCharacterId] = useState('');
  const [character, setCharacter] = useState<CharacterDetail | null>(null);
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [spells, setSpells] = useState<CharacterSpell[]>([]);
  const [activeTab, setActiveTab] = useState<'info' | 'inventory' | 'spells'>('info');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const loadCharacter = async () => {
    if (!characterId) {
      setError('Please enter a character ID');
      return;
    }

    try {
      setLoading(true);
      setError('');

      // Load character details
      const charData = await apiService.getCharacter(characterId);
      setCharacter(charData);

      // Load inventory
      const invData = await apiService.getInventory(characterId);
      setInventory(invData.items || []);

      // Load spells
      const spellData = await apiService.getCharacterSpells(characterId);
      setSpells(spellData.spells || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load character');
    } finally {
      setLoading(false);
    }
  };

  const getHouseColor = (house: string) => {
    const colors: Record<string, string> = {
      gryffindor: 'text-red-700 bg-red-50',
      slytherin: 'text-green-700 bg-green-50',
      hufflepuff: 'text-yellow-700 bg-yellow-50',
      ravenclaw: 'text-blue-700 bg-blue-50',
    };
    return colors[house?.toLowerCase()] || 'text-gray-700 bg-gray-50';
  };

  const getProficiencyLabel = (level: number) => {
    const labels = ['Novice', 'Beginner', 'Intermediate', 'Advanced'];
    return labels[level] || 'Unknown';
  };

  const getProficiencyColor = (level: number) => {
    const colors = [
      'bg-gray-100 text-gray-800',
      'bg-blue-100 text-blue-800',
      'bg-purple-100 text-purple-800',
      'bg-yellow-100 text-yellow-800',
    ];
    return colors[level] || 'bg-gray-100 text-gray-800';
  };

  const getItemTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      currency: 'bg-yellow-100 text-yellow-800',
      balais: 'bg-blue-100 text-blue-800',
      equipment: 'bg-green-100 text-green-800',
      wand: 'bg-purple-100 text-purple-800',
      ingredient: 'bg-orange-100 text-orange-800',
      consumable: 'bg-red-100 text-red-800',
    };
    return colors[type] || 'bg-gray-100 text-gray-800';
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Character Details</h1>

      {/* Search Input */}
      <div className="mb-6 bg-white p-4 rounded-lg shadow">
        <div className="flex gap-2">
          <input
            type="text"
            value={characterId}
            onChange={(e) => setCharacterId(e.target.value)}
            placeholder="Enter Character ID (UUID)"
            className="flex-1 px-3 py-2 border border-gray-300 rounded"
          />
          <button
            onClick={loadCharacter}
            disabled={loading}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            Load Character
          </button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8 text-gray-600">
          Loading character data...
        </div>
      )}

      {/* Character Data */}
      {!loading && character && (
        <>
          {/* Tabs */}
          <div className="mb-4 flex gap-2">
            <button
              onClick={() => setActiveTab('info')}
              className={`px-4 py-2 rounded ${activeTab === 'info' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              Character Info
            </button>
            <button
              onClick={() => setActiveTab('inventory')}
              className={`px-4 py-2 rounded ${activeTab === 'inventory' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              Inventory ({inventory.length})
            </button>
            <button
              onClick={() => setActiveTab('spells')}
              className={`px-4 py-2 rounded ${activeTab === 'spells' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              Spells ({spells.length})
            </button>
          </div>

          {/* Character Info Tab */}
          {activeTab === 'info' && (
            <div className="bg-white rounded-lg shadow p-6">
              <div className="grid grid-cols-2 gap-6">
                <div>
                  <h2 className="text-2xl font-bold mb-4">{character.name}</h2>
                  <div className="space-y-2">
                    <p><strong>Character ID:</strong> {character.id}</p>
                    <p><strong>Account ID:</strong> {character.account_id}</p>
                    <p>
                      <strong>House:</strong>{' '}
                      <span className={`px-2 py-1 rounded font-semibold ${getHouseColor(character.house)}`}>
                        {character.house}
                      </span>
                    </p>
                    <p><strong>Grade (Year):</strong> {character.grade}</p>
                    <p><strong>Level:</strong> {character.level}</p>
                    <p><strong>Experience:</strong> {character.experience.toLocaleString()} XP</p>
                  </div>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-3">Activity</h3>
                  <div className="space-y-2">
                    <p><strong>Created:</strong> {new Date(character.created_at).toLocaleString()}</p>
                  </div>

                  <h3 className="text-lg font-semibold mt-6 mb-3">Statistics</h3>
                  <div className="grid grid-cols-2 gap-2">
                    <div className="bg-blue-50 p-3 rounded text-center">
                      <div className="text-2xl font-bold text-blue-600">{inventory.length}</div>
                      <div className="text-sm text-gray-600">Items</div>
                    </div>
                    <div className="bg-purple-50 p-3 rounded text-center">
                      <div className="text-2xl font-bold text-purple-600">{spells.length}</div>
                      <div className="text-sm text-gray-600">Spells</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Inventory Tab */}
          {activeTab === 'inventory' && (
            <div className="bg-white rounded-lg shadow overflow-hidden">
              {inventory.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  This character has no items in inventory.
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Item</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Quantity</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {inventory.map((item) => (
                      <tr key={item.id}>
                        <td className="px-6 py-4">
                          <div className="text-sm font-medium text-gray-900">{item.display_name}</div>
                          <div className="text-sm text-gray-500">{item.description}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getItemTypeColor(item.item_type)}`}>
                            {item.item_type}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {item.quantity}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                          {item.item_def_id}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}

          {/* Spells Tab */}
          {activeTab === 'spells' && (
            <div className="bg-white rounded-lg shadow overflow-hidden">
              {spells.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  This character has no spells learned.
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Spell</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">School</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Proficiency</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Times Cast</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Learned</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {spells.map((spell) => (
                      <tr key={spell.id}>
                        <td className="px-6 py-4">
                          <div className="text-sm font-medium text-gray-900">{spell.name}</div>
                          <div className="text-sm text-gray-500">{spell.description}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {spell.spell_school}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getProficiencyColor(spell.proficiency_level)}`}>
                            Level {spell.proficiency_level} - {getProficiencyLabel(spell.proficiency_level)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {spell.times_cast}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(spell.learned_at).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}
