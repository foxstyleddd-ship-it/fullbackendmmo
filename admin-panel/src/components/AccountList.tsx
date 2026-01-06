import { useEffect, useState } from 'react';
import { apiService } from '../services/api';

interface Account {
  id: string;
  email: string;
  username: string;
  display_name: string | null;
  discord_id: string | null;
  status: string;
  role: string;
  email_verified: boolean;
  two_factor_enabled: boolean;
  created_at: string;
  last_login_at: string | null;
}

export default function AccountList() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [filteredAccounts, setFilteredAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('all');
  const [filterRole, setFilterRole] = useState('all');
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);

  useEffect(() => {
    loadAccounts();
  }, []);

  useEffect(() => {
    filterAccounts();
  }, [accounts, searchQuery, filterStatus, filterRole]);

  const loadAccounts = async () => {
    try {
      setLoading(true);
      const data = await apiService.getAllAccounts();
      setAccounts(data.accounts || []);
    } catch (err) {
      console.error('Failed to load accounts:', err);
    } finally {
      setLoading(false);
    }
  };

  const filterAccounts = () => {
    let filtered = [...accounts];

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        acc =>
          acc.email.toLowerCase().includes(query) ||
          acc.username.toLowerCase().includes(query) ||
          acc.discord_id?.toLowerCase().includes(query)
      );
    }

    // Apply status filter
    if (filterStatus !== 'all') {
      filtered = filtered.filter(acc => acc.status === filterStatus);
    }

    // Apply role filter
    if (filterRole !== 'all') {
      filtered = filtered.filter(acc => acc.role === filterRole);
    }

    setFilteredAccounts(filtered);
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, string> = {
      active: 'bg-green-100 text-green-800',
      whitelisted: 'bg-blue-100 text-blue-800',
      banned: 'bg-red-100 text-red-800',
      kicked_out: 'bg-yellow-100 text-yellow-800',
    };
    return badges[status] || 'bg-gray-100 text-gray-800';
  };

  const getRoleBadge = (role: string) => {
    const badges: Record<string, string> = {
      player: 'bg-gray-100 text-gray-800',
      gm: 'bg-purple-100 text-purple-800',
      admin: 'bg-indigo-100 text-indigo-800',
      superadmin: 'bg-red-100 text-red-800',
    };
    return badges[role] || 'bg-gray-100 text-gray-800';
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">Account Management</h1>
        <button
          onClick={loadAccounts}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Search</label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Email, username, or Discord ID..."
              className="w-full px-3 py-2 border border-gray-300 rounded"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Status</label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded"
            >
              <option value="all">All Statuses</option>
              <option value="active">Active</option>
              <option value="whitelisted">Whitelisted</option>
              <option value="banned">Banned</option>
              <option value="kicked_out">Kicked Out</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Role</label>
            <select
              value={filterRole}
              onChange={(e) => setFilterRole(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded"
            >
              <option value="all">All Roles</option>
              <option value="player">Player</option>
              <option value="gm">GM</option>
              <option value="admin">Admin</option>
              <option value="superadmin">Superadmin</option>
            </select>
          </div>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8 text-gray-600">Loading accounts...</div>
      )}

      {/* Accounts Table */}
      {!loading && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Username</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Discord ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredAccounts.map((account) => (
                <tr key={account.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{account.username}</div>
                        {account.display_name && typeof account.display_name === 'string' && (
                          <div className="text-sm text-gray-500">{account.display_name}</div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{account.email}</div>
                    {account.email_verified && (
                      <span className="text-xs text-green-600">✓ Verified</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {account.discord_id || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusBadge(account.status)}`}>
                      {account.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getRoleBadge(account.role)}`}>
                      {account.role}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(account.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <button
                      onClick={() => setSelectedAccount(account)}
                      className="text-blue-600 hover:text-blue-900"
                    >
                      View Details
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredAccounts.length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No accounts found matching your filters.
            </div>
          )}
        </div>
      )}

      {/* Stats Summary */}
      {!loading && (
        <div className="mt-6 grid grid-cols-4 gap-4">
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-blue-600">{accounts.length}</div>
            <div className="text-gray-600 text-sm">Total Accounts</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-green-600">
              {accounts.filter(a => a.status === 'active').length}
            </div>
            <div className="text-gray-600 text-sm">Active</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-red-600">
              {accounts.filter(a => a.status === 'banned').length}
            </div>
            <div className="text-gray-600 text-sm">Banned</div>
          </div>
          <div className="bg-white p-4 rounded-lg shadow text-center">
            <div className="text-2xl font-bold text-purple-600">
              {accounts.filter(a => a.role !== 'player').length}
            </div>
            <div className="text-gray-600 text-sm">Staff</div>
          </div>
        </div>
      )}

      {/* Simple Account Detail Modal */}
      {selectedAccount && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full">
            <h2 className="text-2xl font-bold mb-4">Account Details</h2>
            <div className="space-y-2">
              <p><strong>ID:</strong> {selectedAccount.id}</p>
              <p><strong>Username:</strong> {selectedAccount.username}</p>
              {selectedAccount.display_name && typeof selectedAccount.display_name === 'string' && (
                <p><strong>Display Name:</strong> {selectedAccount.display_name}</p>
              )}
              <p><strong>Email:</strong> {selectedAccount.email}</p>
              <p><strong>Discord ID:</strong> {selectedAccount.discord_id || 'Not set'}</p>
              <p><strong>Status:</strong> {selectedAccount.status}</p>
              <p><strong>Role:</strong> {selectedAccount.role}</p>
              <p><strong>2FA:</strong> {selectedAccount.two_factor_enabled ? 'Enabled' : 'Disabled'}</p>
              <p><strong>Created:</strong> {new Date(selectedAccount.created_at).toLocaleString()}</p>
              {selectedAccount.last_login_at && (
                <p><strong>Last Login:</strong> {new Date(selectedAccount.last_login_at).toLocaleString()}</p>
              )}
            </div>
            <button
              onClick={() => setSelectedAccount(null)}
              className="mt-4 px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
