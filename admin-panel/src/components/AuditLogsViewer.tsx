import { useState, useEffect } from 'react';
import { apiService } from '../services/api';
import type { AuditLog } from '../types';

export default function AuditLogsViewer() {
  const [myLogs, setMyLogs] = useState<AuditLog[]>([]);
  const [characterLogs, setCharacterLogs] = useState<AuditLog[]>([]);
  const [characterId, setCharacterId] = useState('');
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<'my' | 'character'>('my');

  useEffect(() => {
    loadMyLogs();
  }, []);

  const loadMyLogs = async () => {
    setLoading(true);
    try {
      const data = await apiService.getMyAuditLogs(100);
      setMyLogs(data.logs);
    } catch (error) {
      console.error('Failed to load audit logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadCharacterLogs = async () => {
    if (!characterId) return;

    setLoading(true);
    try {
      const data = await apiService.getCharacterAuditLogs(characterId, 100);
      setCharacterLogs(data.logs);
    } catch (error) {
      console.error('Failed to load character audit logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const getEventBadgeColor = (eventType: string) => {
    switch (eventType) {
      case 'grade_change': return 'bg-blue-100 text-blue-800';
      case 'teleport': return 'bg-green-100 text-green-800';
      case 'item_grant': return 'bg-purple-100 text-purple-800';
      case 'admin_action': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const renderLog = (log: AuditLog) => (
    <div key={log.id} className="border rounded-lg p-4 hover:shadow-md transition">
      <div className="flex items-start justify-between mb-2">
        <span className={`px-2 py-1 text-xs font-medium rounded ${getEventBadgeColor(log.event_type)}`}>
          {log.event_type}
        </span>
        <span className="text-xs text-gray-500">{formatDate(log.created_at)}</span>
      </div>

      <div className="text-sm font-medium mb-1">{log.action}</div>
      <div className="text-sm text-gray-600 mb-2">{log.details}</div>

      {(log.old_value || log.new_value) && (
        <div className="text-xs text-gray-500 bg-gray-50 rounded p-2 mt-2">
          {log.old_value && <div>Old: {log.old_value}</div>}
          {log.new_value && <div>New: {log.new_value}</div>}
        </div>
      )}

      {log.ip_address && (
        <div className="text-xs text-gray-400 mt-2">IP: {log.ip_address}</div>
      )}

      {!log.success && log.error_message && (
        <div className="text-xs text-red-600 mt-2 bg-red-50 rounded p-2">
          Error: {log.error_message}
        </div>
      )}
    </div>
  );

  return (
    <div className="space-y-6">
      {/* View Toggle */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex gap-4 mb-4">
          <button
            onClick={() => setActiveView('my')}
            className={`px-4 py-2 rounded-lg transition ${
              activeView === 'my'
                ? 'bg-purple-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            My Actions
          </button>
          <button
            onClick={() => setActiveView('character')}
            className={`px-4 py-2 rounded-lg transition ${
              activeView === 'character'
                ? 'bg-purple-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Character History
          </button>
        </div>

        {activeView === 'character' && (
          <div className="flex gap-3">
            <input
              type="text"
              value={characterId}
              onChange={(e) => setCharacterId(e.target.value)}
              placeholder="Enter character ID (UUID)..."
              className="flex-1 px-4 py-2 border rounded-lg"
            />
            <button
              onClick={loadCharacterLogs}
              disabled={loading || !characterId}
              className="px-6 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg disabled:opacity-50"
            >
              Load
            </button>
          </div>
        )}
      </div>

      {/* Logs Display */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold mb-4">
          {activeView === 'my' ? 'My Recent Actions' : 'Character Audit Trail'}
          {loading && <span className="text-sm text-gray-500 ml-2">(Loading...)</span>}
        </h3>

        <div className="space-y-3">
          {activeView === 'my' && myLogs.map(renderLog)}
          {activeView === 'character' && characterLogs.map(renderLog)}

          {activeView === 'my' && myLogs.length === 0 && !loading && (
            <div className="text-center text-gray-500 py-8">
              No audit logs found
            </div>
          )}

          {activeView === 'character' && characterLogs.length === 0 && characterId && !loading && (
            <div className="text-center text-gray-500 py-8">
              No audit logs found for this character
            </div>
          )}

          {activeView === 'character' && !characterId && (
            <div className="text-center text-gray-500 py-8">
              Enter a character ID to view their audit history
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
