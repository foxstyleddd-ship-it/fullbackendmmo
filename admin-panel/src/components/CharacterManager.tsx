import { useState, useEffect } from 'react';
import { apiService } from '../services/api';
import type { CharacterDetail, InventoryItem, SpellDefinition, CharacterSpell } from '../types';

export default function CharacterManager() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCharacter, setSelectedCharacter] = useState<CharacterDetail | null>(null);
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [characterSpells, setCharacterSpells] = useState<CharacterSpell[]>([]);
  const [availableSpells, setAvailableSpells] = useState<SpellDefinition[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{type: 'success' | 'error', text: string} | null>(null);

  // Form states
  const [newGrade, setNewGrade] = useState('');
  const [teleportZone, setTeleportZone] = useState('hogwarts_main');
  const [itemDefId, setItemDefId] = useState('');
  const [itemQuantity, setItemQuantity] = useState('1');
  const [selectedSpellId, setSelectedSpellId] = useState('');

  const zones = [
    'hogwarts_main', 'hogsmeade', 'forbidden_forest', 'quidditch_pitch',
    'ministry_of_magic', 'diagon_alley', 'azkaban'
  ];

  const commonItems = [
    'elder_wand', 'holly_wand', 'healing_potion', 'mana_potion',
    'nimbus_2000', 'firebolt', 'invisibility_cloak', 'felix_felicis'
  ];

  // Load available spells on mount
  useEffect(() => {
    const loadSpells = async () => {
      try {
        const data = await apiService.getAllSpells();
        setAvailableSpells(data.spells);
      } catch (error) {
        console.error('Failed to load spells:', error);
      }
    };
    loadSpells();
  }, []);

  const searchCharacter = async () => {
    if (!searchQuery) return;

    setLoading(true);
    setMessage(null);

    try {
      const characters = await apiService.searchCharacters(searchQuery);
      if (characters.length > 0) {
        loadCharacterDetails(characters[0].id);
      } else {
        setMessage({ type: 'error', text: 'No character found' });
      }
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Search failed' });
    } finally {
      setLoading(false);
    }
  };

  const loadCharacterDetails = async (characterId: string) => {
    setLoading(true);

    try {
      const [character, inventoryData, spellData] = await Promise.all([
        apiService.getCharacter(characterId),
        apiService.getInventory(characterId),
        apiService.getCharacterSpells(characterId),
      ]);

      setSelectedCharacter(character);
      setInventory(inventoryData.items);
      setCharacterSpells(spellData.spells);
      setNewGrade(character.grade.toString());
    } catch (error: any) {
      setMessage({ type: 'error', text: 'Failed to load character details' });
    } finally {
      setLoading(false);
    }
  };

  const handleSetGrade = async () => {
    if (!selectedCharacter) return;

    const grade = parseInt(newGrade);
    if (isNaN(grade) || grade < 1 || grade > 100) {
      setMessage({ type: 'error', text: 'Grade must be between 1 and 100' });
      return;
    }

    setLoading(true);
    try {
      await apiService.setGrade(selectedCharacter.id, grade);
      setMessage({ type: 'success', text: `Grade updated to ${grade}` });
      loadCharacterDetails(selectedCharacter.id);
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Failed to update grade' });
    } finally {
      setLoading(false);
    }
  };

  const handleTeleport = async () => {
    if (!selectedCharacter) return;

    setLoading(true);
    try {
      await apiService.teleport(selectedCharacter.id, teleportZone);
      setMessage({ type: 'success', text: `Teleported to ${teleportZone}` });
      loadCharacterDetails(selectedCharacter.id);
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Teleport failed' });
    } finally {
      setLoading(false);
    }
  };

  const handleGrantItem = async () => {
    if (!selectedCharacter || !itemDefId) return;

    const quantity = parseInt(itemQuantity);
    if (isNaN(quantity) || quantity < 1) {
      setMessage({ type: 'error', text: 'Quantity must be at least 1' });
      return;
    }

    setLoading(true);
    try {
      await apiService.grantItem(selectedCharacter.id, itemDefId, quantity);
      setMessage({ type: 'success', text: `Granted ${quantity}x ${itemDefId}` });

      // Reload inventory
      const inventoryData = await apiService.getInventory(selectedCharacter.id);
      setInventory(inventoryData.items);
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Failed to grant item' });
    } finally {
      setLoading(false);
    }
  };

  const handleGrantSpell = async () => {
    if (!selectedCharacter || !selectedSpellId) return;

    setLoading(true);
    try {
      await apiService.grantSpell(selectedCharacter.id, selectedSpellId);
      const spell = availableSpells.find(s => s.id === selectedSpellId);
      setMessage({ type: 'success', text: `Granted spell: ${spell?.name || selectedSpellId}` });

      // Reload spells
      const spellData = await apiService.getCharacterSpells(selectedCharacter.id);
      setCharacterSpells(spellData.spells);
      setSelectedSpellId('');
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Failed to grant spell' });
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveSpell = async (spellId: string) => {
    if (!selectedCharacter) return;

    setLoading(true);
    try {
      await apiService.removeSpell(selectedCharacter.id, spellId);
      const spell = characterSpells.find(s => s.spell_id === spellId);
      setMessage({ type: 'success', text: `Removed spell: ${spell?.name || spellId}` });

      // Reload spells
      const spellData = await apiService.getCharacterSpells(selectedCharacter.id);
      setCharacterSpells(spellData.spells);
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.error?.message || 'Failed to remove spell' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Search */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">Search Character</h2>
        <div className="flex gap-3">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && searchCharacter()}
            placeholder="Enter character name..."
            className="flex-1 px-4 py-2 border rounded-lg focus:ring-2 focus:ring-purple-500"
          />
          <button
            onClick={searchCharacter}
            disabled={loading}
            className="px-6 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg disabled:opacity-50"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>
      </div>

      {/* Messages */}
      {message && (
        <div className={`rounded-lg p-4 ${
          message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
        }`}>
          {message.text}
        </div>
      )}

      {/* Character Details */}
      {selectedCharacter && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Basic Info */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold mb-4">Character Info</h3>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Name:</span>
                <span className="font-medium">{selectedCharacter.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">House:</span>
                <span className="font-medium">{selectedCharacter.house}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Grade:</span>
                <span className="font-medium">{selectedCharacter.grade}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Level:</span>
                <span className="font-medium">{selectedCharacter.level}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Zone:</span>
                <span className="font-medium">{selectedCharacter.zone_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Galleons:</span>
                <span className="font-medium">{selectedCharacter.currencies?.galleons || 0}</span>
              </div>
            </div>

            {/* Stats */}
            <h4 className="text-md font-semibold mt-6 mb-3">Stats</h4>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>HP: {selectedCharacter.stats.health_current}/{selectedCharacter.stats.health_max}</div>
              <div>MP: {selectedCharacter.stats.mana_current}/{selectedCharacter.stats.mana_max}</div>
              <div>STR: {selectedCharacter.stats.strength}</div>
              <div>DEX: {selectedCharacter.stats.dexterity}</div>
              <div>INT: {selectedCharacter.stats.intelligence}</div>
              <div>WIS: {selectedCharacter.stats.wisdom}</div>
            </div>
          </div>

          {/* GM Actions */}
          <div className="space-y-4">
            {/* Set Grade */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Set Grade</h3>
              <div className="flex gap-3">
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={newGrade}
                  onChange={(e) => setNewGrade(e.target.value)}
                  className="flex-1 px-4 py-2 border rounded-lg"
                  placeholder="1-100"
                />
                <button
                  onClick={handleSetGrade}
                  disabled={loading}
                  className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50"
                >
                  Update
                </button>
              </div>
            </div>

            {/* Teleport */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Teleport</h3>
              <div className="flex gap-3">
                <select
                  value={teleportZone}
                  onChange={(e) => setTeleportZone(e.target.value)}
                  className="flex-1 px-4 py-2 border rounded-lg"
                >
                  {zones.map(zone => (
                    <option key={zone} value={zone}>{zone}</option>
                  ))}
                </select>
                <button
                  onClick={handleTeleport}
                  disabled={loading}
                  className="px-6 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg disabled:opacity-50"
                >
                  Teleport
                </button>
              </div>
            </div>

            {/* Grant Item */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Grant Item</h3>
              <div className="space-y-3">
                <select
                  value={itemDefId}
                  onChange={(e) => setItemDefId(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg"
                >
                  <option value="">Select item...</option>
                  {commonItems.map(item => (
                    <option key={item} value={item}>{item}</option>
                  ))}
                </select>
                <div className="flex gap-3">
                  <input
                    type="number"
                    min="1"
                    value={itemQuantity}
                    onChange={(e) => setItemQuantity(e.target.value)}
                    className="w-24 px-4 py-2 border rounded-lg"
                    placeholder="Qty"
                  />
                  <button
                    onClick={handleGrantItem}
                    disabled={loading || !itemDefId}
                    className="flex-1 px-6 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg disabled:opacity-50"
                  >
                    Grant Item
                  </button>
                </div>
              </div>
            </div>

            {/* Grant Spell */}
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold mb-4">Grant Spell</h3>
              <div className="space-y-3">
                <select
                  value={selectedSpellId}
                  onChange={(e) => setSelectedSpellId(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg text-sm"
                >
                  <option value="">Select spell...</option>
                  {availableSpells.map(spell => (
                    <option key={spell.id} value={spell.id}>
                      {spell.name} ({spell.spell_school}) - Grade {spell.required_grade}
                      {spell.is_forbidden && ' [FORBIDDEN]'}
                    </option>
                  ))}
                </select>
                <button
                  onClick={handleGrantSpell}
                  disabled={loading || !selectedSpellId}
                  className="w-full px-6 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg disabled:opacity-50"
                >
                  Grant Spell
                </button>
              </div>
            </div>
          </div>

          {/* Spells */}
          <div className="bg-white rounded-lg shadow p-6 lg:col-span-2">
            <h3 className="text-lg font-semibold mb-4">Learned Spells ({characterSpells.length} spells)</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {characterSpells.map(spell => (
                <div key={spell.id} className="border rounded-lg p-3 hover:shadow-md transition relative">
                  <div className="font-medium text-sm flex items-center justify-between">
                    <span>{spell.name}</span>
                    {spell.is_forbidden && <span className="text-xs bg-red-600 text-white px-2 py-0.5 rounded">FORBIDDEN</span>}
                  </div>
                  <div className="text-xs text-gray-600 mt-1">
                    {spell.spell_school.replace(/_/g, ' ')}
                  </div>
                  <div className="text-xs mt-2 space-y-1">
                    <div>Mana: {spell.mana_cost} | CD: {spell.cooldown_seconds}s</div>
                    <div>Proficiency: {spell.proficiency_level}/10 | Cast: {spell.times_cast}x</div>
                    <div>Required Grade: {spell.required_grade}</div>
                  </div>
                  <button
                    onClick={() => handleRemoveSpell(spell.spell_id)}
                    disabled={loading}
                    className="mt-2 w-full px-3 py-1 bg-red-500 hover:bg-red-600 text-white text-xs rounded disabled:opacity-50"
                  >
                    Remove
                  </button>
                </div>
              ))}
              {characterSpells.length === 0 && (
                <div className="text-gray-500 text-sm col-span-full text-center py-4">
                  No spells learned
                </div>
              )}
            </div>
          </div>

          {/* Inventory */}
          <div className="bg-white rounded-lg shadow p-6 lg:col-span-2">
            <h3 className="text-lg font-semibold mb-4">Inventory ({inventory.length} items)</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {inventory.map(item => (
                <div key={item.id} className="border rounded-lg p-3 hover:shadow-md transition">
                  <div className="font-medium text-sm">{item.display_name}</div>
                  <div className="text-xs text-gray-600">
                    {item.item_type} {item.equipment_slot && `(${item.equipment_slot})`}
                  </div>
                  <div className="text-xs mt-1">
                    Qty: {item.quantity} | Value: {item.base_value}g
                  </div>
                </div>
              ))}
              {inventory.length === 0 && (
                <div className="text-gray-500 text-sm col-span-full text-center py-4">
                  No items in inventory
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {!selectedCharacter && !loading && (
        <div className="text-center text-gray-500 py-12">
          Search for a character to begin
        </div>
      )}
    </div>
  );
}
