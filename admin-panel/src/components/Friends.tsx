import { useState } from 'react';
import { apiService } from '../services/api';

interface Friend {
  friendship_id: string;
  character_id: string;
  character_name: string;
  house: string;
  grade: number;
  level: number;
  status: string;
  is_requester: boolean;
  last_played_at: string | null;
  created_at: string;
}

export default function Friends() {
  const [viewMode, setViewMode] = useState<'friends' | 'requests'>('friends');
  const [characterId, setCharacterId] = useState('');
  const [friends, setFriends] = useState<Friend[]>([]);
  const [requests, setRequests] = useState<Friend[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  // Send friend request fields
  const [requesterCharId, setRequesterCharId] = useState('');
  const [addresseeCharId, setAddresseeCharId] = useState('');

  const loadFriends = async () => {
    if (!characterId) {
      setError('Please enter a character ID');
      return;
    }

    try {
      setLoading(true);
      setError('');
      setMessage(null);
      const data = await apiService.getFriends(characterId);
      setFriends(data.friends || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load friends');
    } finally {
      setLoading(false);
    }
  };

  const loadPendingRequests = async () => {
    if (!characterId) {
      setError('Please enter a character ID');
      return;
    }

    try {
      setLoading(true);
      setError('');
      setMessage(null);
      const data = await apiService.getPendingRequests(characterId);
      setRequests(data.requests || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to load requests');
    } finally {
      setLoading(false);
    }
  };

  const sendFriendRequest = async () => {
    if (!requesterCharId || !addresseeCharId) {
      setMessage({ type: 'error', text: 'Both character IDs are required' });
      return;
    }

    try {
      setLoading(true);
      setMessage(null);
      await apiService.sendFriendRequest(requesterCharId, addresseeCharId);
      setMessage({ type: 'success', text: 'Friend request sent!' });
      setRequesterCharId('');
      setAddresseeCharId('');
    } catch (err: any) {
      setMessage({ type: 'error', text: err.response?.data?.error?.message || 'Failed to send request' });
    } finally {
      setLoading(false);
    }
  };

  const acceptRequest = async (friendshipId: string) => {
    try {
      setLoading(true);
      setMessage(null);
      await apiService.acceptFriendRequest(friendshipId);
      setMessage({ type: 'success', text: 'Friend request accepted!' });
      loadPendingRequests();
    } catch (err: any) {
      setMessage({ type: 'error', text: err.response?.data?.error?.message || 'Failed to accept request' });
    } finally {
      setLoading(false);
    }
  };

  const declineRequest = async (friendshipId: string) => {
    try {
      setLoading(true);
      setMessage(null);
      await apiService.declineFriendRequest(friendshipId);
      setMessage({ type: 'success', text: 'Friend request declined' });
      loadPendingRequests();
    } catch (err: any) {
      setMessage({ type: 'error', text: err.response?.data?.error?.message || 'Failed to decline request' });
    } finally {
      setLoading(false);
    }
  };

  const removeFriend = async (friendshipId: string, friendName: string) => {
    if (!confirm(`Remove ${friendName} from friends?`)) return;

    try {
      setLoading(true);
      setMessage(null);
      await apiService.removeFriend(friendshipId);
      setMessage({ type: 'success', text: `${friendName} removed from friends` });
      loadFriends();
    } catch (err: any) {
      setMessage({ type: 'error', text: err.response?.data?.error?.message || 'Failed to remove friend' });
    } finally {
      setLoading(false);
    }
  };

  const getHouseColor = (house: string) => {
    const colors: Record<string, string> = {
      gryffindor: 'text-red-700',
      slytherin: 'text-green-700',
      hufflepuff: 'text-yellow-700',
      ravenclaw: 'text-blue-700',
    };
    return colors[house.toLowerCase()] || 'text-gray-700';
  };

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Friends Management</h1>

      {/* Send Friend Request Section */}
      <div className="mb-6 bg-white p-4 rounded-lg shadow">
        <h2 className="text-xl font-semibold mb-3">Send Friend Request</h2>
        <div className="grid grid-cols-2 gap-3">
          <input
            type="text"
            value={requesterCharId}
            onChange={(e) => setRequesterCharId(e.target.value)}
            placeholder="Requester Character ID"
            className="px-3 py-2 border border-gray-300 rounded"
          />
          <input
            type="text"
            value={addresseeCharId}
            onChange={(e) => setAddresseeCharId(e.target.value)}
            placeholder="Addressee Character ID"
            className="px-3 py-2 border border-gray-300 rounded"
          />
        </div>
        <button
          onClick={sendFriendRequest}
          disabled={loading}
          className="mt-3 px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:bg-gray-400"
        >
          Send Request
        </button>
      </div>

      {/* View Mode Toggle */}
      <div className="mb-6 flex gap-2">
        <button
          onClick={() => setViewMode('friends')}
          className={`px-4 py-2 rounded ${viewMode === 'friends' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          Friends List
        </button>
        <button
          onClick={() => setViewMode('requests')}
          className={`px-4 py-2 rounded ${viewMode === 'requests' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          Pending Requests
        </button>
      </div>

      {/* Character ID Input */}
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
            onClick={viewMode === 'friends' ? loadFriends : loadPendingRequests}
            disabled={loading}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            Load
          </button>
        </div>
      </div>

      {/* Messages */}
      {message && (
        <div className={`mb-4 p-3 rounded border ${message.type === 'success' ? 'bg-green-100 border-green-400 text-green-700' : 'bg-red-100 border-red-400 text-red-700'}`}>
          {message.text}
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
          Loading...
        </div>
      )}

      {/* Friends List */}
      {viewMode === 'friends' && !loading && (
        <div className="space-y-3">
          {friends.length === 0 && characterId && (
            <div className="text-center py-8 text-gray-600">
              No friends found for this character.
            </div>
          )}
          {friends.map(friend => (
            <div key={friend.friendship_id} className="bg-white p-4 rounded-lg shadow hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-lg font-semibold text-gray-900">{friend.character_name}</h3>
                    <span className={`font-semibold ${getHouseColor(friend.house)}`}>
                      {friend.house}
                    </span>
                    <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">
                      Year {friend.grade}
                    </span>
                    <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-sm">
                      Level {friend.level}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600">
                    <span>Character ID: {friend.character_id}</span>
                    {friend.last_played_at && (
                      <span className="ml-4">
                        Last played: {new Date(friend.last_played_at).toLocaleString()}
                      </span>
                    )}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    Friends since: {new Date(friend.created_at).toLocaleDateString()}
                  </div>
                </div>
                <button
                  onClick={() => removeFriend(friend.friendship_id, friend.character_name)}
                  disabled={loading}
                  className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:bg-gray-400"
                >
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pending Requests */}
      {viewMode === 'requests' && !loading && (
        <div className="space-y-3">
          {requests.length === 0 && characterId && (
            <div className="text-center py-8 text-gray-600">
              No pending friend requests.
            </div>
          )}
          {requests.map(request => (
            <div key={request.friendship_id} className="bg-white p-4 rounded-lg shadow hover:shadow-md transition-shadow border-l-4 border-yellow-400">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-lg font-semibold text-gray-900">{request.character_name}</h3>
                    <span className={`font-semibold ${getHouseColor(request.house)}`}>
                      {request.house}
                    </span>
                    <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">
                      Year {request.grade}
                    </span>
                    <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded text-sm">
                      Level {request.level}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600">
                    Character ID: {request.character_id}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    Requested: {new Date(request.created_at).toLocaleString()}
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => acceptRequest(request.friendship_id)}
                    disabled={loading}
                    className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:bg-gray-400"
                  >
                    Accept
                  </button>
                  <button
                    onClick={() => declineRequest(request.friendship_id)}
                    disabled={loading}
                    className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:bg-gray-400"
                  >
                    Decline
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
