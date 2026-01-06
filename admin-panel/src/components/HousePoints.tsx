import { useState } from 'react';
import { apiService } from '../services/api';

export default function HousePoints() {
  const [action, setAction] = useState<'add' | 'remove' | 'reset'>('add');
  const [house, setHouse] = useState('gryffindor');
  const [points, setPoints] = useState('');
  const [characterName, setCharacterName] = useState('');
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  const houses = [
    { value: 'gryffindor', label: 'Gryffindor', color: 'text-red-700', bg: 'bg-red-50' },
    { value: 'slytherin', label: 'Slytherin', color: 'text-green-700', bg: 'bg-green-50' },
    { value: 'hufflepuff', label: 'Hufflepuff', color: 'text-yellow-700', bg: 'bg-yellow-50' },
    { value: 'ravenclaw', label: 'Ravenclaw', color: 'text-blue-700', bg: 'bg-blue-50' },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!points || Number(points) <= 0) {
      setMessage({ type: 'error', text: 'Please enter a valid number of points' });
      return;
    }

    if (!reason.trim()) {
      setMessage({ type: 'error', text: 'Please provide a reason' });
      return;
    }

    try {
      setLoading(true);
      setMessage(null);

      const command = action === 'add' ? 'addhousepoints' : 'removehousepoints';

      await apiService.executeCommand({
        command,
        target_id: house,
        parameters: {
          points: Number(points),
          reason: reason.trim(),
          character_name: characterName.trim() || null,
        },
      });

      setMessage({
        type: 'success',
        text: `Successfully ${action === 'add' ? 'added' : 'removed'} ${points} points ${action === 'add' ? 'to' : 'from'} ${house}!`,
      });

      // Reset form
      setPoints('');
      setCharacterName('');
      setReason('');
    } catch (err: any) {
      setMessage({
        type: 'error',
        text: err.response?.data?.error?.message || 'Failed to update house points',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">House Points Management</h1>

      <div className="bg-white rounded-lg shadow p-6">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Action Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Action</label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setAction('add')}
                className={`px-4 py-2 rounded ${action === 'add' ? 'bg-green-600 text-white' : 'bg-gray-200 text-gray-700'}`}
              >
                Add Points
              </button>
              <button
                type="button"
                onClick={() => setAction('remove')}
                className={`px-4 py-2 rounded ${action === 'remove' ? 'bg-red-600 text-white' : 'bg-gray-200 text-gray-700'}`}
              >
                Remove Points
              </button>
            </div>
          </div>

          {/* House Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">House</label>
            <div className="grid grid-cols-4 gap-2">
              {houses.map(h => (
                <button
                  key={h.value}
                  type="button"
                  onClick={() => setHouse(h.value)}
                  className={`px-4 py-3 rounded font-semibold transition ${
                    house === h.value
                      ? `${h.bg} ${h.color} border-2 border-current`
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {h.label}
                </button>
              ))}
            </div>
          </div>

          {/* Points Input */}
          <div>
            <label htmlFor="points" className="block text-sm font-medium text-gray-700 mb-2">
              Points {action === 'add' ? 'to Add' : 'to Remove'}
            </label>
            <input
              id="points"
              type="number"
              min="1"
              value={points}
              onChange={(e) => setPoints(e.target.value)}
              placeholder="Enter number of points"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          {/* Character Name (Optional) */}
          <div>
            <label htmlFor="characterName" className="block text-sm font-medium text-gray-700 mb-2">
              Character Name (Optional)
            </label>
            <input
              id="characterName"
              type="text"
              value={characterName}
              onChange={(e) => setCharacterName(e.target.value)}
              placeholder="Leave empty for global house points"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
            />
            <p className="mt-1 text-sm text-gray-500">
              If specified, this will be recorded in the audit log. Leave empty to apply to entire house.
            </p>
          </div>

          {/* Reason */}
          <div>
            <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-2">
              Reason
            </label>
            <textarea
              id="reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Explain why points are being added/removed..."
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          {/* Submit Button */}
          <div>
            <button
              type="submit"
              disabled={loading}
              className={`w-full px-6 py-3 rounded font-semibold text-white ${
                action === 'add'
                  ? 'bg-green-600 hover:bg-green-700'
                  : 'bg-red-600 hover:bg-red-700'
              } disabled:bg-gray-400 transition`}
            >
              {loading
                ? 'Processing...'
                : `${action === 'add' ? 'Add' : 'Remove'} ${points || '0'} Points ${action === 'add' ? 'to' : 'from'} ${house}`}
            </button>
          </div>

          {/* Message Display */}
          {message && (
            <div
              className={`p-4 rounded border ${
                message.type === 'success'
                  ? 'bg-green-100 border-green-400 text-green-700'
                  : 'bg-red-100 border-red-400 text-red-700'
              }`}
            >
              {message.text}
            </div>
          )}
        </form>

        {/* Info Box */}
        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded">
          <h3 className="font-semibold text-blue-900 mb-2">How it works:</h3>
          <ul className="list-disc list-inside text-sm text-blue-800 space-y-1">
            <li>Adding/removing points affects ALL characters in the selected house</li>
            <li>Character name is optional - used only for audit trail</li>
            <li>Reason is required for accountability</li>
            <li>All actions are logged in the audit system</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
