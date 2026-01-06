import { useState } from 'react';
import { apiService } from '../services/api';

export default function CreateCharacter() {
  const [name, setName] = useState('');
  const [house, setHouse] = useState('no_house');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{type: 'success' | 'error', text: string} | null>(null);

  const houses = [
    { value: 'no_house', label: 'Pas de maison', color: 'text-gray-700' },
    { value: 'brumval', label: 'Brumval', color: 'text-red-700' },
    { value: 'aerwyn', label: 'Aerwyn', color: 'text-green-700' },
    { value: 'falcon', label: 'Falcon', color: 'text-yellow-700' },
    { value: 'venatrix', label: 'Venatrix', color: 'text-blue-700' },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name || !house) {
      setMessage({ type: 'error', text: 'All fields are required' });
      return;
    }

    setLoading(true);
    setMessage(null);

    try {
      const character = await apiService.createCharacter({
        name,
        house,
        appearance_data: {}
      });

      setMessage({
        type: 'success',
        text: `Character "${character.name}" created successfully! House: ${character.house}`
      });

      // Reset form
      setName('');
      setHouse('no_house');
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.error?.message || 'Failed to create character'
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold mb-6">Create New Character</h2>

        {message && (
          <div
            className={`mb-6 p-4 rounded-lg ${
              message.type === 'success'
                ? 'bg-green-50 text-green-800 border border-green-200'
                : 'bg-red-50 text-red-800 border border-red-200'
            }`}
          >
            {message.text}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
              Character Name
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              placeholder="Harry Potter"
              minLength={3}
              maxLength={64}
              pattern="[a-zA-Z][a-zA-Z\s']{2,63}"
              required
            />
            <p className="mt-1 text-sm text-gray-500">
              3-64 characters, letters, spaces, and apostrophes only
            </p>
          </div>

          <div>
            <label htmlFor="house" className="block text-sm font-medium text-gray-700 mb-2">
              House
            </label>
            <select
              id="house"
              value={house}
              onChange={(e) => setHouse(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              required
            >
              {houses.map((h) => (
                <option key={h.value} value={h.value}>
                  {h.label}
                </option>
              ))}
            </select>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full px-6 py-3 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Creating Character...' : 'Create Character'}
          </button>
        </form>

        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="font-medium text-blue-900 mb-2">Character Details:</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• Starting Level: 1</li>
            <li>• Starting Grade: 1</li>
            <li>• Starting Zone: Hogwarts Main</li>
            <li>• You can modify stats, spells, and inventory after creation</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
