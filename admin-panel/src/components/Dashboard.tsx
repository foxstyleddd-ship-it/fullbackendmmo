import { useState } from 'react';
import { apiService } from '../services/api';
import CharacterManager from './CharacterManager';
import AuditLogsViewer from './AuditLogsViewer';
import CreateAccount from './CreateAccount';
import CreateCharacter from './CreateCharacter';
import Achievements from './Achievements';
import Friends from './Friends';
import Leaderboard from './Leaderboard';
import AccountList from './AccountList';
import CharacterDetail from './CharacterDetail';
import HousePoints from './HousePoints';
import Notebooks from './Notebooks';
import GradeReport from './GradeReport';
import ItemDefinitions from './ItemDefinitions';
import InventoryManager from './InventoryManager';

interface DashboardProps {
  onLogout: () => void;
}

type Tab = 'characters' | 'character-detail' | 'accounts' | 'create-account' | 'create-character' | 'achievements' | 'friends' | 'leaderboard' | 'house-points' | 'notebooks' | 'grades' | 'audit' | 'items' | 'inventory';

export default function Dashboard({ onLogout }: DashboardProps) {
  const [activeTab, setActiveTab] = useState<Tab>('characters');
  const user = apiService.getCurrentUser();

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">🎮 HP MMO Admin Panel</h1>
              <p className="text-sm text-gray-600">
                Logged in as: <span className="font-medium">{user?.username}</span> ({user?.role})
              </p>
            </div>
            <button
              onClick={onLogout}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('characters')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'characters'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Character Management
            </button>
            <button
              onClick={() => setActiveTab('create-account')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'create-account'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Create Account
            </button>
            <button
              onClick={() => setActiveTab('create-character')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'create-character'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Create Character
            </button>
            <button
              onClick={() => setActiveTab('character-detail')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'character-detail'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Character Detail
            </button>
            <button
              onClick={() => setActiveTab('accounts')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'accounts'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Accounts
            </button>
            <button
              onClick={() => setActiveTab('achievements')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'achievements'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Achievements
            </button>
            <button
              onClick={() => setActiveTab('friends')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'friends'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Friends
            </button>
            <button
              onClick={() => setActiveTab('leaderboard')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'leaderboard'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Leaderboard
            </button>
            <button
              onClick={() => setActiveTab('house-points')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'house-points'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              House Points
            </button>
            <button
              onClick={() => setActiveTab('notebooks')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'notebooks'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Carnets
            </button>
            <button
              onClick={() => setActiveTab('grades')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'grades'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Notes & Bulletins
            </button>
            <button
              onClick={() => setActiveTab('audit')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'audit'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Audit Logs
            </button>
            <button
              onClick={() => setActiveTab('items')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'items'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Item Definitions
            </button>
            <button
              onClick={() => setActiveTab('inventory')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition ${
                activeTab === 'inventory'
                  ? 'border-purple-500 text-purple-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Inventory Manager
            </button>
          </nav>
        </div>
      </div>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'characters' && <CharacterManager />}
        {activeTab === 'character-detail' && <CharacterDetail />}
        {activeTab === 'accounts' && <AccountList />}
        {activeTab === 'create-account' && <CreateAccount />}
        {activeTab === 'create-character' && <CreateCharacter />}
        {activeTab === 'achievements' && <Achievements />}
        {activeTab === 'friends' && <Friends />}
        {activeTab === 'leaderboard' && <Leaderboard />}
        {activeTab === 'house-points' && <HousePoints />}
        {activeTab === 'notebooks' && <Notebooks />}
        {activeTab === 'grades' && <GradeReport />}
        {activeTab === 'audit' && <AuditLogsViewer />}
        {activeTab === 'items' && <ItemDefinitions />}
        {activeTab === 'inventory' && <InventoryManager />}
      </main>
    </div>
  );
}
