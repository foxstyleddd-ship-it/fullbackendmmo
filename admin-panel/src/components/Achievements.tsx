import { useEffect, useState } from 'react';
import { apiService } from '../services/api';

interface Achievement {
  id: string;
  name: string;
  description: string;
  category: string;
  icon_path: string;
  points: number;
  is_secret: boolean;
  required_count: number;
  reward_currencies: Record<string, number>;
}

interface CharacterAchievement {
  id: string;
  character_id: string;
  achievement_id: string;
  progress: number;
  unlocked_at: string | null;
  name: string;
  description: string;
  category: string;
  icon_path: string;
  points: number;
  is_secret: boolean;
  required_count: number;
}

interface AchievementStats {
  total: number;
  unlocked: number;
  points: number;
}

export default function Achievements() {
  const [viewMode, setViewMode] = useState<'all' | 'character'>('all');
  const [allAchievements, setAllAchievements] = useState<Achievement[]>([]);
  const [characterAchievements, setCharacterAchievements] = useState<CharacterAchievement[]>([]);
  const [stats, setStats] = useState<AchievementStats | null>(null);
  const [selectedCharacterId, setSelectedCharacterId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<string>('all');

  const categories = ['all', 'combat', 'exploration', 'social', 'collection', 'progression'];

  useEffect(() => {
    loadAllAchievements();
  }, []);

  const loadAllAchievements = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await apiService.getAllAchievements();
      setAllAchievements(data.achievements || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load achievements');
    } finally {
      setLoading(false);
    }
  };

  const loadCharacterAchievements = async () => {
    if (!selectedCharacterId) {
      setError('Please enter a character ID');
      return;
    }

    try {
      setLoading(true);
      setError('');
      const data = await apiService.getCharacterAchievements(selectedCharacterId);
      setCharacterAchievements(data.achievements || []);
      setStats(data.stats || null);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load character achievements');
    } finally {
      setLoading(false);
    }
  };

  const filteredAchievements = allAchievements.filter(a =>
    filter === 'all' || a.category === filter
  );

  const filteredCharacterAchievements = characterAchievements.filter(a =>
    filter === 'all' || a.category === filter
  );

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      combat: 'bg-red-100 text-red-800',
      exploration: 'bg-green-100 text-green-800',
      social: 'bg-blue-100 text-blue-800',
      collection: 'bg-purple-100 text-purple-800',
      progression: 'bg-yellow-100 text-yellow-800',
    };
    return colors[category] || 'bg-gray-100 text-gray-800';
  };

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Achievements</h1>

      {/* View Mode Toggle */}
      <div className="mb-6 flex gap-2">
        <button
          onClick={() => setViewMode('all')}
          className={`px-4 py-2 rounded ${viewMode === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          All Achievements
        </button>
        <button
          onClick={() => setViewMode('character')}
          className={`px-4 py-2 rounded ${viewMode === 'character' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          Character Progress
        </button>
      </div>

      {/* Character ID Input (only shown in character view) */}
      {viewMode === 'character' && (
        <div className="mb-6 bg-white p-4 rounded-lg shadow">
          <div className="flex gap-2">
            <input
              type="text"
              value={selectedCharacterId}
              onChange={(e) => setSelectedCharacterId(e.target.value)}
              placeholder="Enter Character ID (UUID)"
              className="flex-1 px-3 py-2 border border-gray-300 rounded"
            />
            <button
              onClick={loadCharacterAchievements}
              disabled={loading}
              className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
            >
              Load
            </button>
          </div>
        </div>
      )}

      {/* Stats Display (character view only) */}
      {viewMode === 'character' && stats && (
        <div className="mb-6 grid grid-cols-3 gap-4">
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-3xl font-bold text-blue-600">{stats.unlocked}</div>
            <div className="text-gray-600">Unlocked</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-3xl font-bold text-gray-600">{stats.total}</div>
            <div className="text-gray-600">Total</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-3xl font-bold text-yellow-600">{stats.points}</div>
            <div className="text-gray-600">Points</div>
          </div>
        </div>
      )}

      {/* Category Filter */}
      <div className="mb-4 flex gap-2 flex-wrap">
        {categories.map(cat => (
          <button
            key={cat}
            onClick={() => setFilter(cat)}
            className={`px-3 py-1 rounded text-sm ${filter === cat ? 'bg-gray-800 text-white' : 'bg-gray-200 text-gray-700'}`}
          >
            {cat.charAt(0).toUpperCase() + cat.slice(1)}
          </button>
        ))}
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
          Loading achievements...
        </div>
      )}

      {/* All Achievements View */}
      {viewMode === 'all' && !loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredAchievements.map(achievement => (
            <div key={achievement.id} className="bg-white p-4 rounded-lg shadow hover:shadow-md transition-shadow">
              <div className="flex items-start gap-3">
                <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center text-2xl">
                  🏆
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <h3 className="font-semibold text-gray-900">{achievement.name}</h3>
                    <span className={`px-2 py-1 rounded text-xs ${getCategoryColor(achievement.category)}`}>
                      {achievement.category}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 mb-2">{achievement.description}</p>
                  <div className="flex items-center gap-4 text-xs text-gray-500">
                    <span className="font-semibold text-yellow-600">{achievement.points} pts</span>
                    {achievement.required_count > 1 && (
                      <span>Required: {achievement.required_count}</span>
                    )}
                    {achievement.is_secret && (
                      <span className="text-purple-600">Secret</span>
                    )}
                  </div>
                  {Object.keys(achievement.reward_currencies).length > 0 && (
                    <div className="mt-2 text-xs text-green-600">
                      Rewards: {JSON.stringify(achievement.reward_currencies)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Character Achievements View */}
      {viewMode === 'character' && !loading && characterAchievements.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredCharacterAchievements.map(achievement => {
            const progress = Math.min(100, (achievement.progress / achievement.required_count) * 100);
            const isUnlocked = achievement.unlocked_at !== null;

            return (
              <div key={achievement.id} className={`bg-white p-4 rounded-lg shadow hover:shadow-md transition-shadow ${isUnlocked ? 'border-2 border-yellow-400' : ''}`}>
                <div className="flex items-start gap-3">
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center text-2xl ${isUnlocked ? 'bg-yellow-200' : 'bg-gray-100'}`}>
                    {isUnlocked ? '✅' : '🔒'}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold text-gray-900">{achievement.name}</h3>
                      <span className={`px-2 py-1 rounded text-xs ${getCategoryColor(achievement.category)}`}>
                        {achievement.category}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{achievement.description}</p>

                    {/* Progress Bar */}
                    {!isUnlocked && (
                      <div className="mb-2">
                        <div className="flex justify-between text-xs text-gray-600 mb-1">
                          <span>Progress: {achievement.progress} / {achievement.required_count}</span>
                          <span>{progress.toFixed(0)}%</span>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-2">
                          <div
                            className="bg-blue-600 h-2 rounded-full transition-all"
                            style={{ width: `${progress}%` }}
                          />
                        </div>
                      </div>
                    )}

                    <div className="flex items-center gap-4 text-xs text-gray-500">
                      <span className="font-semibold text-yellow-600">{achievement.points} pts</span>
                      {isUnlocked && (
                        <span className="text-green-600">
                          Unlocked: {new Date(achievement.unlocked_at!).toLocaleDateString()}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Empty State */}
      {viewMode === 'character' && !loading && characterAchievements.length === 0 && selectedCharacterId && (
        <div className="text-center py-8 text-gray-600">
          No achievements found for this character.
        </div>
      )}
    </div>
  );
}
