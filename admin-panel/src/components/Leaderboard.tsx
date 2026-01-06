import { useEffect, useState } from 'react';
import { apiService } from '../services/api';

interface LeaderboardEntry {
  id: string;
  name: string;
  house: string;
  grade: number;
  level: number;
  experience: number;
  house_points: number;
  achievement_count: number;
  total_playtime_seconds: number;
  last_played_at: string | null;
}

export default function Leaderboard() {
  const [viewMode, setViewMode] = useState<'overall' | 'house'>('overall');
  const [selectedHouse, setSelectedHouse] = useState('gryffindor');
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [limit, setLimit] = useState(50);

  const houses = [
    { value: 'gryffindor', label: 'Gryffindor', color: 'text-red-700', bg: 'bg-red-50' },
    { value: 'slytherin', label: 'Slytherin', color: 'text-green-700', bg: 'bg-green-50' },
    { value: 'hufflepuff', label: 'Hufflepuff', color: 'text-yellow-700', bg: 'bg-yellow-50' },
    { value: 'ravenclaw', label: 'Ravenclaw', color: 'text-blue-700', bg: 'bg-blue-50' },
  ];

  useEffect(() => {
    loadLeaderboard();
  }, [viewMode, selectedHouse, limit]);

  const loadLeaderboard = async () => {
    try {
      setLoading(true);
      setError('');
      const data = viewMode === 'overall'
        ? await apiService.getLeaderboard(limit)
        : await apiService.getHouseLeaderboard(selectedHouse, limit);
      setEntries(data.leaderboard || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load leaderboard');
    } finally {
      setLoading(false);
    }
  };

  const formatPlaytime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  const getMedalEmoji = (rank: number) => {
    if (rank === 1) return '🥇';
    if (rank === 2) return '🥈';
    if (rank === 3) return '🥉';
    return `#${rank}`;
  };

  const getHouseInfo = (houseName: string) => {
    return houses.find(h => h.value === houseName.toLowerCase()) || houses[0];
  };

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Leaderboard</h1>

      {/* View Mode Toggle */}
      <div className="mb-6 flex gap-2 items-center">
        <button
          onClick={() => setViewMode('overall')}
          className={`px-4 py-2 rounded ${viewMode === 'overall' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          Overall Leaderboard
        </button>
        <button
          onClick={() => setViewMode('house')}
          className={`px-4 py-2 rounded ${viewMode === 'house' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          House Leaderboard
        </button>

        {/* Limit Selector */}
        <select
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          className="ml-auto px-3 py-2 border border-gray-300 rounded"
        >
          <option value={10}>Top 10</option>
          <option value={25}>Top 25</option>
          <option value={50}>Top 50</option>
          <option value={100}>Top 100</option>
        </select>
      </div>

      {/* House Selector (only shown in house view) */}
      {viewMode === 'house' && (
        <div className="mb-6 flex gap-2">
          {houses.map(house => (
            <button
              key={house.value}
              onClick={() => setSelectedHouse(house.value)}
              className={`px-4 py-2 rounded font-semibold ${selectedHouse === house.value ? `${house.bg} ${house.color} border-2 border-current` : 'bg-gray-100 text-gray-700'}`}
            >
              {house.label}
            </button>
          ))}
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8 text-gray-600">
          Loading leaderboard...
        </div>
      )}

      {/* Leaderboard Table */}
      {!loading && entries.length > 0 && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Rank
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Character
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  House
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Level
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Experience
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  House Points
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Achievements
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Playtime
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {entries.map((entry, index) => {
                const rank = index + 1;
                const houseInfo = getHouseInfo(entry.house);
                const isTopThree = rank <= 3;

                return (
                  <tr key={entry.id} className={isTopThree ? houseInfo.bg : ''}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-lg font-bold text-gray-900">
                        {getMedalEmoji(rank)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{entry.name}</div>
                      <div className="text-xs text-gray-500">Year {entry.grade}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`font-semibold ${houseInfo.color}`}>
                        {entry.house}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{entry.level}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{entry.experience.toLocaleString()}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                        {entry.house_points}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-purple-100 text-purple-800">
                        {entry.achievement_count}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatPlaytime(entry.total_playtime_seconds)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Empty State */}
      {!loading && entries.length === 0 && (
        <div className="text-center py-8 text-gray-600 bg-white rounded-lg shadow">
          No entries found in the leaderboard.
        </div>
      )}

      {/* Stats Summary */}
      {!loading && entries.length > 0 && (
        <div className="mt-6 grid grid-cols-4 gap-4">
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-blue-600">{entries.length}</div>
            <div className="text-gray-600 text-sm">Players</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-green-600">
              {entries[0]?.level || 0}
            </div>
            <div className="text-gray-600 text-sm">Top Level</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-yellow-600">
              {entries[0]?.house_points || 0}
            </div>
            <div className="text-gray-600 text-sm">Top House Points</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-purple-600">
              {entries[0]?.achievement_count || 0}
            </div>
            <div className="text-gray-600 text-sm">Top Achievements</div>
          </div>
        </div>
      )}
    </div>
  );
}
