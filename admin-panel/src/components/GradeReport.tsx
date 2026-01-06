import { useState } from 'react';
import { apiService } from '../services/api';

interface Grade {
  id: string;
  character_id: string;
  school_year: number;
  subject: string;
  grade_value: string;
  exam_type: string;
  score: number | null;
  teacher_comment: string | null;
  awarded_at: string;
}

interface ReportCard {
  id: string;
  character_id: string;
  school_year: number;
  overall_comment: string | null;
  headmaster_signature: string | null;
  issued_at: string;
  grades: Grade[];
}

export default function GradeReport() {
  const [characterId, setCharacterId] = useState('');
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [grades, setGrades] = useState<Grade[]>([]);
  const [reportCards, setReportCards] = useState<ReportCard[]>([]);
  const [viewMode, setViewMode] = useState<'grades' | 'reports'>('grades');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const subjects = {
    charms: 'Enchantements',
    transfiguration: 'Métamorphose',
    potions: 'Potions',
    defense_against_dark_arts: 'Défense contre les Forces du Mal',
    herbology: 'Botanique',
    astronomy: 'Astronomie',
    history_of_magic: 'Histoire de la Magie',
    care_of_magical_creatures: 'Soins aux Créatures Magiques',
    divination: 'Divination',
    ancient_runes: 'Runes Anciennes',
    arithmancy: 'Arithmancie',
    muggle_studies: 'Étude des Moldus',
    flying: 'Vol',
  };

  const gradeValues = {
    outstanding: { label: 'Optimal (O)', color: 'bg-green-100 text-green-800' },
    exceeds_expectations: { label: 'Effort Exceptionnel (E)', color: 'bg-blue-100 text-blue-800' },
    acceptable: { label: 'Acceptable (A)', color: 'bg-yellow-100 text-yellow-800' },
    poor: { label: 'Piètre (P)', color: 'bg-orange-100 text-orange-800' },
    dreadful: { label: 'Désolant (D)', color: 'bg-red-100 text-red-800' },
    troll: { label: 'Troll (T)', color: 'bg-gray-100 text-gray-800' },
  };

  const examTypes = {
    continuous: 'Contrôle Continu',
    midterm: 'Examen de Mi-Année',
    final: 'Examen Final',
    owls: 'BUSES',
    newts: 'ASPICS',
  };

  const loadGrades = async () => {
    if (!characterId) {
      setError('Veuillez entrer un ID de personnage');
      return;
    }

    try {
      setLoading(true);
      setError('');
      const data = await apiService.getCharacterGrades(characterId, selectedYear ?? undefined);
      setGrades(data.grades || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Échec du chargement des notes');
    } finally {
      setLoading(false);
    }
  };

  const loadReportCards = async () => {
    if (!characterId) {
      setError('Veuillez entrer un ID de personnage');
      return;
    }

    try {
      setLoading(true);
      setError('');
      const data = await apiService.getAllReportCards(characterId);
      setReportCards(data.report_cards || []);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Échec du chargement des bulletins');
    } finally {
      setLoading(false);
    }
  };

  const handleLoad = () => {
    if (viewMode === 'grades') {
      loadGrades();
    } else {
      loadReportCards();
    }
  };

  const getSubjectLabel = (key: string) => subjects[key as keyof typeof subjects] || key;
  const getGradeInfo = (key: string) => gradeValues[key as keyof typeof gradeValues] || { label: key, color: 'bg-gray-100 text-gray-800' };
  const getExamType = (key: string) => examTypes[key as keyof typeof examTypes] || key;

  const groupGradesByYear = (grades: Grade[]) => {
    const grouped: Record<number, Grade[]> = {};
    grades.forEach(grade => {
      if (!grouped[grade.school_year]) {
        grouped[grade.school_year] = [];
      }
      grouped[grade.school_year].push(grade);
    });
    return grouped;
  };

  const groupedGrades = groupGradesByYear(grades);

  return (
    <div className="max-w-7xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">📚 Notes et Bulletins</h1>

      {/* Controls */}
      <div className="mb-6 bg-white p-4 rounded-lg shadow">
        <div className="flex gap-2 mb-4">
          <button
            onClick={() => setViewMode('grades')}
            className={`px-4 py-2 rounded ${viewMode === 'grades' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
          >
            Notes par Matière
          </button>
          <button
            onClick={() => setViewMode('reports')}
            className={`px-4 py-2 rounded ${viewMode === 'reports' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
          >
            Bulletins de Fin d'Année
          </button>
        </div>

        <div className="flex gap-2">
          <input
            type="text"
            value={characterId}
            onChange={(e) => setCharacterId(e.target.value)}
            placeholder="ID du personnage (UUID)"
            className="flex-1 px-3 py-2 border border-gray-300 rounded"
          />
          {viewMode === 'grades' && (
            <select
              value={selectedYear || ''}
              onChange={(e) => setSelectedYear(e.target.value ? Number(e.target.value) : null)}
              className="px-3 py-2 border border-gray-300 rounded"
            >
              <option value="">Toutes les années</option>
              {[1, 2, 3, 4, 5, 6, 7].map(year => (
                <option key={year} value={year}>Année {year}</option>
              ))}
            </select>
          )}
          <button
            onClick={handleLoad}
            disabled={loading}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            Charger
          </button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8 text-gray-600">Chargement...</div>
      )}

      {/* Grades View */}
      {viewMode === 'grades' && !loading && (
        <div className="space-y-6">
          {Object.entries(groupedGrades).sort(([a], [b]) => Number(b) - Number(a)).map(([year, yearGrades]) => (
            <div key={year} className="bg-white rounded-lg shadow overflow-hidden">
              <div className="bg-indigo-600 text-white px-6 py-3">
                <h2 className="text-xl font-bold">Année {year}</h2>
              </div>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Matière</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Note</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Score</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Commentaire</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {yearGrades.map((grade) => {
                      const gradeInfo = getGradeInfo(grade.grade_value);
                      return (
                        <tr key={grade.id}>
                          <td className="px-6 py-4 whitespace-nowrap font-medium text-gray-900">
                            {getSubjectLabel(grade.subject)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${gradeInfo.color}`}>
                              {gradeInfo.label}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                            {getExamType(grade.exam_type)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {grade.score !== null ? `${grade.score}/100` : '-'}
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-600">
                            {grade.teacher_comment || '-'}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {new Date(grade.awarded_at).toLocaleDateString()}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          ))}

          {grades.length === 0 && characterId && (
            <div className="text-center py-12 bg-white rounded-lg shadow">
              <p className="text-gray-500">Aucune note trouvée pour ce personnage</p>
            </div>
          )}
        </div>
      )}

      {/* Report Cards View */}
      {viewMode === 'reports' && !loading && (
        <div className="space-y-6">
          {reportCards.map((report) => (
            <div key={report.id} className="bg-white rounded-lg shadow-lg overflow-hidden border-4 border-indigo-200">
              {/* Report Header */}
              <div className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white px-8 py-6">
                <div className="text-center">
                  <h2 className="text-3xl font-bold mb-2">🎓 Bulletin Scolaire</h2>
                  <p className="text-xl">Année {report.school_year}</p>
                  <p className="text-sm opacity-90">Émis le {new Date(report.issued_at).toLocaleDateString()}</p>
                </div>
              </div>

              {/* Grades Table */}
              <div className="p-6">
                <h3 className="text-lg font-semibold mb-4">Notes par Matière</h3>
                <div className="overflow-x-auto mb-6">
                  <table className="min-w-full divide-y divide-gray-200 border">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Matière</th>
                        <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Note</th>
                        <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Commentaire</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {report.grades.map((grade) => {
                        const gradeInfo = getGradeInfo(grade.grade_value);
                        return (
                          <tr key={grade.id}>
                            <td className="px-4 py-3 font-medium text-gray-900">
                              {getSubjectLabel(grade.subject)}
                            </td>
                            <td className="px-4 py-3">
                              <span className={`px-2 py-1 inline-flex text-xs font-semibold rounded-full ${gradeInfo.color}`}>
                                {gradeInfo.label}
                              </span>
                            </td>
                            <td className="px-4 py-3 text-sm text-gray-600">
                              {grade.teacher_comment || '-'}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Overall Comment */}
                {report.overall_comment && (
                  <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-4">
                    <h4 className="font-semibold text-gray-900 mb-2">Remarque Générale du Directeur</h4>
                    <p className="text-gray-700 italic">{report.overall_comment}</p>
                  </div>
                )}

                {/* Signature */}
                {report.headmaster_signature && (
                  <div className="text-right mt-6">
                    <p className="text-gray-600">Signature du Directeur:</p>
                    <p className="font-cursive text-2xl text-indigo-700">{report.headmaster_signature}</p>
                  </div>
                )}
              </div>
            </div>
          ))}

          {reportCards.length === 0 && characterId && (
            <div className="text-center py-12 bg-white rounded-lg shadow">
              <p className="text-gray-500">Aucun bulletin trouvé pour ce personnage</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
