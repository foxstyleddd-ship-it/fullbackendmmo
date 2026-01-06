import { useState, useEffect } from 'react';
import { apiService } from '../services/api';
import type { ItemDefinition, ItemType, EquipmentSlotType } from '../types';

export default function ItemDefinitions() {
  const [items, setItems] = useState<ItemDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingItem, setEditingItem] = useState<ItemDefinition | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const [formData, setFormData] = useState<Partial<ItemDefinition>>({
    id: '',
    item_type: 'consumable',
    display_name: '',
    description: '',
    icon_path: '',
    is_stackable: true,
    max_stack_size: 99,
    is_tradeable: true,
    is_droppable: true,
    is_destroyable: true,
    required_grade: 1,
    required_level: 1,
    properties: {},
    base_value: 0,
  });

  useEffect(() => {
    loadItems();
  }, []);

  const loadItems = async () => {
    try {
      setLoading(true);
      const data = await apiService.getAllItemDefinitions();
      setItems(data.items || []);
    } catch (error: any) {
      setMessage({ type: 'error', text: 'Failed to load items' });
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingItem) {
        await apiService.updateItemDefinition(editingItem.id, formData);
        setMessage({ type: 'success', text: 'Item updated successfully!' });
      } else {
        await apiService.createItemDefinition(formData);
        setMessage({ type: 'success', text: 'Item created successfully!' });
      }
      setShowModal(false);
      setEditingItem(null);
      resetForm();
      loadItems();
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error?.message || 'Failed to save item'
      });
    }
  };

  const handleEdit = (item: ItemDefinition) => {
    setEditingItem(item);
    setFormData(item);
    setShowModal(true);
  };

  const handleDelete = async (itemId: string) => {
    if (!confirm('Are you sure you want to delete this item?')) return;

    try {
      await apiService.deleteItemDefinition(itemId);
      setMessage({ type: 'success', text: 'Item deleted successfully!' });
      loadItems();
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error?.message || 'Failed to delete item'
      });
    }
  };

  const resetForm = () => {
    setFormData({
      id: '',
      item_type: 'consumable',
      display_name: '',
      description: '',
      icon_path: '',
      is_stackable: true,
      max_stack_size: 99,
      is_tradeable: true,
      is_droppable: true,
      is_destroyable: true,
      required_grade: 1,
      required_level: 1,
      properties: {},
      base_value: 0,
    });
  };

  const filteredItems = items.filter(item =>
    item.display_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    item.id.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const itemTypes: ItemType[] = ['weapon', 'armor', 'accessory', 'consumable', 'quest_item', 'material', 'cosmetic', 'currency', 'broom', 'pet'];
  const equipmentSlots: EquipmentSlotType[] = ['wand', 'robe', 'hat', 'cloak', 'amulet', 'ring_left', 'ring_right', 'boots', 'gloves', 'broom'];

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Item Definitions</h1>
        <button
          onClick={() => {
            resetForm();
            setEditingItem(null);
            setShowModal(true);
          }}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Create New Item
        </button>
      </div>

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

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search items..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {loading ? (
        <div className="text-center py-8">Loading...</div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Value</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Stackable</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredItems.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-600">{item.id}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{item.display_name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{item.item_type}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{item.base_value}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                    {item.is_stackable ? `Yes (${item.max_stack_size})` : 'No'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      onClick={() => handleEdit(item)}
                      className="text-blue-600 hover:text-blue-900 mr-4"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(item.id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {filteredItems.length === 0 && (
            <div className="text-center py-8 text-gray-500">No items found</div>
          )}
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto p-6">
            <h2 className="text-2xl font-bold mb-4">
              {editingItem ? 'Edit Item' : 'Create New Item'}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Item ID *
                  </label>
                  <input
                    type="text"
                    required
                    disabled={!!editingItem}
                    value={formData.id}
                    onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Display Name *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.display_name}
                    onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Item Type *</label>
                  <select
                    required
                    value={formData.item_type}
                    onChange={(e) => setFormData({ ...formData, item_type: e.target.value as ItemType })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  >
                    {itemTypes.map((type) => (
                      <option key={type} value={type}>{type}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Equipment Slot</label>
                  <select
                    value={formData.equipment_slot || ''}
                    onChange={(e) => setFormData({ ...formData, equipment_slot: e.target.value as EquipmentSlotType || undefined })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">None</option>
                    {equipmentSlots.map((slot) => (
                      <option key={slot} value={slot}>{slot}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Base Value</label>
                  <input
                    type="number"
                    value={formData.base_value}
                    onChange={(e) => setFormData({ ...formData, base_value: parseInt(e.target.value) })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Required Grade</label>
                  <input
                    type="number"
                    value={formData.required_grade}
                    onChange={(e) => setFormData({ ...formData, required_grade: parseInt(e.target.value) })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Required Level</label>
                  <input
                    type="number"
                    value={formData.required_level}
                    onChange={(e) => setFormData({ ...formData, required_level: parseInt(e.target.value) })}
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="flex items-center space-x-6">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_stackable}
                    onChange={(e) => setFormData({ ...formData, is_stackable: e.target.checked })}
                    className="mr-2"
                  />
                  Stackable
                </label>
                {formData.is_stackable && (
                  <div>
                    <label className="text-sm">Max Stack:</label>
                    <input
                      type="number"
                      value={formData.max_stack_size}
                      onChange={(e) => setFormData({ ...formData, max_stack_size: parseInt(e.target.value) })}
                      className="ml-2 w-20 px-2 py-1 border border-gray-300 rounded"
                    />
                  </div>
                )}
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_tradeable}
                    onChange={(e) => setFormData({ ...formData, is_tradeable: e.target.checked })}
                    className="mr-2"
                  />
                  Tradeable
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_droppable}
                    onChange={(e) => setFormData({ ...formData, is_droppable: e.target.checked })}
                    className="mr-2"
                  />
                  Droppable
                </label>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Icon Path</label>
                <input
                  type="text"
                  value={formData.icon_path}
                  onChange={(e) => setFormData({ ...formData, icon_path: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingItem(null);
                    resetForm();
                  }}
                  className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                  {editingItem ? 'Update' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
