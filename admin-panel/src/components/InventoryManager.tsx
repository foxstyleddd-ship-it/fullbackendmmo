import { useState, useEffect } from 'react';
import { apiService } from '../services/api';
import type { InventoryItemWithDef, ItemDefinition, Character } from '../types';

export default function InventoryManager() {
  const [characters, setCharacters] = useState<Character[]>([]);
  const [selectedCharacter, setSelectedCharacter] = useState<string>('');
  const [inventory, setInventory] = useState<InventoryItemWithDef[]>([]);
  const [allItems, setAllItems] = useState<ItemDefinition[]>([]);
  const [loading, setLoading] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  const [addForm, setAddForm] = useState({
    item_def_id: '',
    quantity: 1
  });

  useEffect(() => {
    loadCharacters();
    loadAllItems();
  }, []);

  useEffect(() => {
    if (selectedCharacter) {
      loadInventory();
    }
  }, [selectedCharacter]);

  const loadCharacters = async () => {
    try {
      const data = await apiService.searchCharacters('');
      setCharacters(data);
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to load characters' });
    }
  };

  const loadAllItems = async () => {
    try {
      const data = await apiService.getAllItemDefinitions();
      setAllItems(data.items || []);
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to load items' });
    }
  };

  const loadInventory = async () => {
    if (!selectedCharacter) return;

    try {
      setLoading(true);
      const data = await apiService.getCharacterInventoryAdmin(selectedCharacter);
      setInventory(data.items || []);
    } catch (error: any) {
      setMessage({ type: 'error', text: 'Failed to load inventory' });
    } finally {
      setLoading(false);
    }
  };

  const handleAddItem = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedCharacter) return;

    try {
      await apiService.addItemToInventory(selectedCharacter, addForm.item_def_id, addForm.quantity);
      setMessage({ type: 'success', text: 'Item added successfully!' });
      setShowAddModal(false);
      setAddForm({ item_def_id: '', quantity: 1 });
      loadInventory();
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error?.message || 'Failed to add item'
      });
    }
  };

  const handleRemoveItem = async (itemDefId: string, maxQuantity: number) => {
    const quantity = prompt(`How many to remove? (Max: ${maxQuantity})`);
    if (!quantity || parseInt(quantity) <= 0) return;

    const qty = Math.min(parseInt(quantity), maxQuantity);

    try {
      await apiService.removeItemFromInventory(selectedCharacter, itemDefId, qty);
      setMessage({ type: 'success', text: 'Item removed successfully!' });
      loadInventory();
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error?.message || 'Failed to remove item'
      });
    }
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Inventory Manager</h1>

      {message && (
        <div
          className={`mb-4 p-4 rounded ${
            message.type === 'success'
              ? 'bg-green-100 text-green-800 border border-green-200'
              : 'bg-red-100 text-red-800 border border-red-200'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Select Character
        </label>
        <select
          value={selectedCharacter}
          onChange={(e) => setSelectedCharacter(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
        >
          <option value="">-- Select a character --</option>
          {characters.map((char) => (
            <option key={char.id} value={char.id}>
              {char.name} (Level {char.level}, Grade {char.grade})
            </option>
          ))}
        </select>
      </div>

      {selectedCharacter && (
        <>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">
              Inventory ({inventory.length} items)
            </h2>
            <button
              onClick={() => setShowAddModal(true)}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
            >
              Add Item
            </button>
          </div>

          {loading ? (
            <div className="text-center py-8">Loading inventory...</div>
          ) : (
            <div className="bg-white rounded-lg shadow overflow-hidden">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Item</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Quantity</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Value</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Requirements</th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {inventory.map((item) => (
                    <tr key={item.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div className="text-sm font-medium text-gray-900">{item.display_name}</div>
                        <div className="text-sm text-gray-500">{item.description}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        {item.item_type}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        {item.quantity}
                        {item.is_stackable && ` / ${item.max_stack_size}`}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        {item.base_value}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                        Grade {item.required_grade}, Level {item.required_level}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleRemoveItem(item.item_def_id, item.quantity)}
                          className="text-red-600 hover:text-red-900"
                        >
                          Remove
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {inventory.length === 0 && (
                <div className="text-center py-8 text-gray-500">No items in inventory</div>
              )}
            </div>
          )}
        </>
      )}

      {/* Add Item Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h2 className="text-2xl font-bold mb-4">Add Item to Inventory</h2>
            <form onSubmit={handleAddItem} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Item *
                </label>
                <select
                  required
                  value={addForm.item_def_id}
                  onChange={(e) => setAddForm({ ...addForm, item_def_id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">-- Select an item --</option>
                  {allItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.display_name} ({item.item_type})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Quantity *
                </label>
                <input
                  type="number"
                  required
                  min="1"
                  value={addForm.quantity}
                  onChange={(e) => setAddForm({ ...addForm, quantity: parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false);
                    setAddForm({ item_def_id: '', quantity: 1 });
                  }}
                  className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                >
                  Add Item
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
