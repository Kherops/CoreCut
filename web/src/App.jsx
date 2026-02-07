import React, { useState, useEffect } from 'react';
import {
  LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
  AreaChart, Area
} from 'recharts';
import { 
  Upload, TrendingUp, TrendingDown, Server, CheckCircle, AlertCircle, FileJson, 
  DollarSign, Zap, Leaf, Clock, Calculator, Settings, Sun, Moon, ChevronDown,
  BarChart3, Activity, Cpu, HardDrive
} from 'lucide-react';

// Default cost parameters (user can adjust)
const DEFAULT_COST_PARAMS = {
  hourlyRate: 50,
  electricityPrice: 0.25,
  serverWatts: 150,
  executionsPerDay: 1000,
  workingDaysPerYear: 250,
};

// Theme configuration
const themes = {
  light: {
    bg: 'bg-gray-50',
    bgSecondary: 'bg-white',
    bgTertiary: 'bg-gray-100',
    text: 'text-gray-900',
    textSecondary: 'text-gray-600',
    textMuted: 'text-gray-400',
    border: 'border-gray-200',
    borderHover: 'hover:border-gray-300',
    card: 'bg-white shadow-sm border border-gray-200',
    cardHover: 'hover:shadow-md hover:border-gray-300',
    input: 'bg-white border-gray-300 text-gray-900',
    accent: 'bg-indigo-600 hover:bg-indigo-700',
    accentText: 'text-indigo-600',
    success: 'text-emerald-600',
    successBg: 'bg-emerald-50 border-emerald-200',
    danger: 'text-red-600',
    dangerBg: 'bg-red-50 border-red-200',
    warning: 'text-amber-600',
    warningBg: 'bg-amber-50 border-amber-200',
    chartGrid: '#e5e7eb',
    chartText: '#6b7280',
  },
  dark: {
    bg: 'bg-gray-950',
    bgSecondary: 'bg-gray-900',
    bgTertiary: 'bg-gray-800',
    text: 'text-white',
    textSecondary: 'text-gray-300',
    textMuted: 'text-gray-500',
    border: 'border-gray-800',
    borderHover: 'hover:border-gray-700',
    card: 'bg-gray-900 border border-gray-800',
    cardHover: 'hover:border-gray-700',
    input: 'bg-gray-800 border-gray-700 text-white',
    accent: 'bg-indigo-500 hover:bg-indigo-600',
    accentText: 'text-indigo-400',
    success: 'text-emerald-400',
    successBg: 'bg-emerald-950/50 border-emerald-800',
    danger: 'text-red-400',
    dangerBg: 'bg-red-950/50 border-red-800',
    warning: 'text-amber-400',
    warningBg: 'bg-amber-950/50 border-amber-800',
    chartGrid: '#374151',
    chartText: '#9ca3af',
  }
};

function App() {
  const [reportData, setReportData] = useState(null);
  const [aggregateData, setAggregateData] = useState(null);
  const [activeTab, setActiveTab] = useState('single');
  const [error, setError] = useState(null);
  const [costParams, setCostParams] = useState(DEFAULT_COST_PARAMS);
  const [showCostSettings, setShowCostSettings] = useState(false);
  const [isDark, setIsDark] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('corecut-theme') === 'dark' || 
        (!localStorage.getItem('corecut-theme') && window.matchMedia('(prefers-color-scheme: dark)').matches);
    }
    return true;
  });

  const theme = isDark ? themes.dark : themes.light;

  useEffect(() => {
    localStorage.setItem('corecut-theme', isDark ? 'dark' : 'light');
  }, [isDark]);

  const handleFileUpload = (event, type) => {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const json = JSON.parse(e.target.result);
        if (type === 'single') {
          setReportData(json);
          setActiveTab('single');
        } else {
          setAggregateData(json);
          setActiveTab('aggregate');
        }
        setError(null);
      } catch (err) {
        setError('Invalid JSON file');
      }
    };
    reader.readAsText(file);
  };

  const loadSampleData = () => {
    const sampleReport = {
      version: "1.0",
      generated_at: new Date().toISOString(),
      machine: "demo-machine",
      tag: "v1.0.0",
      config: {
        baseline_script: "./baseline.sh",
        optimized_script: "./optimized.sh",
        mode: "duration",
        warmup_runs: 1,
        measured_runs: 9,
        alternate: true,
        cooldown_ms: 500,
        timeout: 300
      },
      baseline: {
        runs: [
          { duration_ms: 1250, exit_code: 0 },
          { duration_ms: 1180, exit_code: 0 },
          { duration_ms: 1320, exit_code: 0 },
          { duration_ms: 1200, exit_code: 0 },
          { duration_ms: 1280, exit_code: 0 },
          { duration_ms: 1150, exit_code: 0 },
          { duration_ms: 1300, exit_code: 0 },
          { duration_ms: 1220, exit_code: 0 },
          { duration_ms: 1190, exit_code: 0 }
        ],
        stats: { median: 1220, mean: 1232, std_dev: 55, cv: 4.5, p10: 1160, p90: 1310 }
      },
      optimized: {
        runs: [
          { duration_ms: 850, exit_code: 0 },
          { duration_ms: 820, exit_code: 0 },
          { duration_ms: 880, exit_code: 0 },
          { duration_ms: 840, exit_code: 0 },
          { duration_ms: 860, exit_code: 0 },
          { duration_ms: 810, exit_code: 0 },
          { duration_ms: 870, exit_code: 0 },
          { duration_ms: 830, exit_code: 0 },
          { duration_ms: 845, exit_code: 0 }
        ],
        stats: { median: 845, mean: 845, std_dev: 22, cv: 2.6, p10: 815, p90: 875 }
      },
      comparison: {
        gain_percent: 30.74,
        gain_p10: 28.5,
        gain_p90: 33.2,
        conclusive: true,
        overlap: 0.05
      }
    };
    setReportData(sampleReport);
    setActiveTab('single');
  };

  return (
    <div className={`min-h-screen transition-colors duration-300 ${theme.bg}`}>
      {/* Header - Clean & Modern */}
      <header className={`sticky top-0 z-50 backdrop-blur-xl ${theme.bgSecondary}/80 border-b ${theme.border}`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo & Brand */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-3">
                <div className="relative">
                  <img 
                    src="/CoreCut.png" 
                    alt="CoreCut" 
                    className="h-10 w-10 object-contain rounded-lg shadow-lg"
                  />
                  <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-emerald-500 rounded-full border-2 border-white dark:border-gray-900"></div>
                </div>
                <div>
                  <h1 className={`text-xl font-bold ${theme.text}`}>CoreCut</h1>
                  <p className={`text-xs ${theme.textMuted}`}>Performance Analytics</p>
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2">
              {/* Theme Toggle */}
              <button
                onClick={() => setIsDark(!isDark)}
                className={`p-2.5 rounded-xl ${theme.bgTertiary} ${theme.text} hover:scale-105 transition-all duration-200`}
                title={isDark ? 'Mode clair' : 'Mode sombre'}
              >
                {isDark ? <Sun size={18} /> : <Moon size={18} />}
              </button>

              {/* Cost Settings */}
              <button
                onClick={() => setShowCostSettings(!showCostSettings)}
                className={`p-2.5 rounded-xl ${showCostSettings ? 'bg-amber-500 text-white' : `${theme.bgTertiary} ${theme.text}`} hover:scale-105 transition-all duration-200`}
                title="Paramètres de coûts"
              >
                <Calculator size={18} />
              </button>

              <div className="w-px h-6 bg-gray-300 dark:bg-gray-700 mx-1"></div>

              {/* File Actions */}
              <label className={`cursor-pointer px-4 py-2 rounded-xl ${theme.accent} text-white font-medium flex items-center gap-2 transition-all duration-200 hover:scale-105 shadow-lg shadow-indigo-500/25`}>
                <Upload size={16} />
                <span className="hidden sm:inline">Charger</span>
                <input type="file" accept=".json" className="hidden" onChange={(e) => handleFileUpload(e, 'single')} />
              </label>

              <button
                onClick={loadSampleData}
                className={`px-4 py-2 rounded-xl ${theme.bgTertiary} ${theme.text} font-medium transition-all duration-200 hover:scale-105`}
              >
                Démo
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Cost Settings Panel - Collapsible */}
      {showCostSettings && (
        <div className={`border-b ${theme.border} ${theme.bgSecondary}`}>
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <div className="flex items-center gap-2 mb-4">
              <Settings size={18} className={theme.warning} />
              <h3 className={`font-semibold ${theme.text}`}>Paramètres de Calcul des Coûts</h3>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
              <CostInput
                label="Coût horaire (€/h)"
                value={costParams.hourlyRate}
                onChange={(v) => setCostParams({...costParams, hourlyRate: v})}
                theme={theme}
              />
              <CostInput
                label="Prix élec. (€/kWh)"
                value={costParams.electricityPrice}
                onChange={(v) => setCostParams({...costParams, electricityPrice: v})}
                theme={theme}
              />
              <CostInput
                label="Puissance (W)"
                value={costParams.serverWatts}
                onChange={(v) => setCostParams({...costParams, serverWatts: v})}
                theme={theme}
              />
              <CostInput
                label="Exécutions/jour"
                value={costParams.executionsPerDay}
                onChange={(v) => setCostParams({...costParams, executionsPerDay: v})}
                theme={theme}
              />
              <CostInput
                label="Jours/an"
                value={costParams.workingDaysPerYear}
                onChange={(v) => setCostParams({...costParams, workingDaysPerYear: v})}
                tooltip="Nombre de jours ouvrés par an"
              />
            </div>
          </div>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className={`border-b ${theme.border} ${theme.bgSecondary}`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav className="flex gap-1 py-2">
            <button
              onClick={() => setActiveTab('single')}
              className={`px-4 py-2.5 rounded-lg font-medium text-sm transition-all duration-200 flex items-center gap-2 ${
                activeTab === 'single'
                  ? `${theme.accent} text-white shadow-lg`
                  : `${theme.textSecondary} hover:${theme.bgTertiary}`
              }`}
            >
              <Cpu size={16} />
              Machine Unique
            </button>
            <button
              onClick={() => setActiveTab('aggregate')}
              className={`px-4 py-2.5 rounded-lg font-medium text-sm transition-all duration-200 flex items-center gap-2 ${
                activeTab === 'aggregate'
                  ? `${theme.accent} text-white shadow-lg`
                  : `${theme.textSecondary} hover:${theme.bgTertiary}`
              }`}
            >
              <HardDrive size={16} />
              Multi-Machines
            </button>
          </nav>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
          <div className={`${theme.dangerBg} border px-4 py-3 rounded-xl flex items-center gap-2`}>
            <AlertCircle size={18} className={theme.danger} />
            <span className={theme.danger}>{error}</span>
          </div>
        </div>
      )}

      {/* Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'single' && (
          reportData ? <SingleReport data={reportData} costParams={costParams} theme={theme} /> : <EmptyState type="single" theme={theme} />
        )}
        {activeTab === 'aggregate' && (
          aggregateData ? <AggregateReport data={aggregateData} costParams={costParams} theme={theme} /> : <EmptyState type="aggregate" theme={theme} />
        )}
      </main>

      {/* Footer */}
      <footer className={`border-t ${theme.border} py-6 mt-auto`}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <p className={`text-sm ${theme.textMuted}`}>
            CoreCut — Les comparaisons sont relatives (A vs B sur la même machine).
          </p>
          <p className={`text-xs ${theme.textMuted} mt-1`}>
            Les temps bruts ne sont jamais comparés entre machines différentes.
          </p>
        </div>
      </footer>
    </div>
  );
}

function EmptyState({ type, theme }) {
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <div className={`p-6 rounded-2xl ${theme.bgTertiary} mb-6`}>
        <FileJson size={48} className={theme.textMuted} />
      </div>
      <h2 className={`text-xl font-semibold mb-2 ${theme.text}`}>
        {type === 'single' ? 'Aucun rapport chargé' : 'Aucun agrégat chargé'}
      </h2>
      <p className={`text-sm ${theme.textMuted} text-center max-w-md`}>
        {type === 'single'
          ? 'Chargez un fichier report_*.json ou cliquez sur "Démo" pour voir un exemple'
          : 'Chargez un fichier aggregate.json pour voir les résultats multi-machines'}
      </p>
      <div className={`mt-8 flex items-center gap-4 text-sm ${theme.textMuted}`}>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-indigo-500"></div>
          <span>Glissez-déposez un fichier</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-emerald-500"></div>
          <span>Ou utilisez le bouton Charger</span>
        </div>
      </div>
    </div>
  );
}

function SingleReport({ data, costParams, theme }) {
  const gainPositive = data.comparison.gain_percent >= 0;
  
  const chartData = data.baseline.runs.map((run, i) => ({
    name: `#${i + 1}`,
    baseline: run.duration_ms,
    optimized: data.optimized.runs[i]?.duration_ms || 0
  }));

  // Calculate financial impact
  const baselineMs = data.baseline.stats.median;
  const optimizedMs = data.optimized.stats.median;
  const timeSavedMs = baselineMs - optimizedMs;
  const timeSavedPerExec = timeSavedMs / 1000;
  
  const dailyTimeSavedSec = timeSavedPerExec * costParams.executionsPerDay;
  const dailyTimeSavedHours = dailyTimeSavedSec / 3600;
  const annualTimeSavedHours = dailyTimeSavedHours * costParams.workingDaysPerYear;
  const annualCostSavings = annualTimeSavedHours * costParams.hourlyRate;
  
  const baselineEnergyPerExec = (baselineMs / 1000 / 3600) * costParams.serverWatts;
  const optimizedEnergyPerExec = (optimizedMs / 1000 / 3600) * costParams.serverWatts;
  const energySavedPerExec = baselineEnergyPerExec - optimizedEnergyPerExec;
  
  const dailyEnergySavedWh = energySavedPerExec * costParams.executionsPerDay;
  const annualEnergySavedKwh = (dailyEnergySavedWh * costParams.workingDaysPerYear) / 1000;
  const annualElectricitySavings = annualEnergySavedKwh * costParams.electricityPrice;
  
  const co2PerKwh = 0.4;
  const annualCo2SavedKg = annualEnergySavedKwh * co2PerKwh;
  const totalAnnualSavings = annualCostSavings + annualElectricitySavings;

  return (
    <div className="space-y-6">
      {/* Hero Gain Card */}
      <div className={`${theme.card} rounded-2xl p-8 text-center`}>
        <div className="flex items-center justify-center gap-2 mb-4">
          <Activity size={20} className={theme.accentText} />
          <h2 className={`text-lg font-semibold ${theme.textSecondary}`}>Gain de Performance</h2>
        </div>
        <div className={`text-6xl font-bold mb-4 flex items-center justify-center gap-3 ${
          gainPositive ? theme.success : theme.danger
        }`}>
          {gainPositive ? <TrendingUp size={48} /> : <TrendingDown size={48} />}
          {data.comparison.gain_percent.toFixed(2)}%
        </div>
        <div className={`flex items-center justify-center gap-4 text-sm ${theme.textSecondary}`}>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-indigo-500"></div>
            <span>Baseline: <strong className={theme.text}>{baselineMs.toFixed(0)}ms</strong></span>
          </div>
          <div className={`w-px h-4 ${theme.border}`}></div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-emerald-500"></div>
            <span>Optimisé: <strong className={theme.text}>{optimizedMs.toFixed(0)}ms</strong></span>
          </div>
        </div>
        <div className="mt-6">
          {data.comparison.conclusive ? (
            <span className={`inline-flex items-center gap-2 px-4 py-2 rounded-full ${theme.successBg} ${theme.success} text-sm font-medium`}>
              <CheckCircle size={16} /> Résultat concluant
            </span>
          ) : (
            <span className={`inline-flex items-center gap-2 px-4 py-2 rounded-full ${theme.warningBg} ${theme.warning} text-sm font-medium`}>
              <AlertCircle size={16} /> Non concluant (variance élevée)
            </span>
          )}
        </div>
      </div>

      {/* Machine Info */}
      <div className={`${theme.card} rounded-xl p-4 flex items-center justify-between`}>
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg ${theme.bgTertiary}`}>
            <Server size={18} className={theme.textMuted} />
          </div>
          <div>
            <span className={`font-medium ${theme.text}`}>{data.machine}</span>
            {data.tag && <span className={`ml-2 text-sm ${theme.textMuted}`}>({data.tag})</span>}
          </div>
        </div>
        <div className={`text-sm ${theme.textMuted}`}>
          {new Date(data.generated_at).toLocaleString('fr-FR')}
        </div>
      </div>

      {/* Financial Impact Section */}
      <div className={`${theme.card} rounded-2xl p-6 border-2 border-emerald-500/20`}>
        <div className="flex items-center gap-3 mb-6">
          <div className="p-3 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/25">
            <DollarSign size={24} className="text-white" />
          </div>
          <div>
            <h2 className={`text-xl font-bold ${theme.text}`}>Impact Financier & Environnemental</h2>
            <p className={`text-sm ${theme.textMuted}`}>Ce que cette optimisation vous fait gagner</p>
          </div>
        </div>

        {/* Big Money Number */}
        <div className="text-center mb-8">
          <p className={`text-sm mb-2 ${theme.textMuted}`}>Économies annuelles estimées</p>
          <div className={`text-5xl font-bold ${totalAnnualSavings >= 0 ? theme.success : theme.danger}`}>
            {totalAnnualSavings >= 0 ? '+' : ''}{totalAnnualSavings.toLocaleString('fr-FR', { minimumFractionDigits: 0, maximumFractionDigits: 0 })} €
          </div>
          <p className={`text-sm mt-2 ${theme.textMuted}`}>par an</p>
        </div>

        {/* Impact Cards Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
          <ImpactCard icon={Clock} label="Temps économisé" value={`${timeSavedMs.toFixed(0)} ms`} subValue={`${annualTimeSavedHours.toFixed(1)}h /an`} color="blue" theme={theme} />
          <ImpactCard icon={DollarSign} label="Coût serveur" value={`+${annualCostSavings.toFixed(0)} €`} subValue={`${(annualCostSavings / 12).toFixed(0)} € /mois`} color="emerald" theme={theme} />
          <ImpactCard icon={Zap} label="Électricité" value={`${annualEnergySavedKwh.toFixed(1)} kWh`} subValue={`+${annualElectricitySavings.toFixed(2)} € /an`} color="amber" theme={theme} />
          <ImpactCard icon={Leaf} label="CO₂ évité" value={`-${annualCo2SavedKg.toFixed(1)} kg`} subValue={`${(annualCo2SavedKg / 12).toFixed(2)} kg /mois`} color="green" theme={theme} />
        </div>

        {/* Before/After Comparison */}
        <div className="grid md:grid-cols-2 gap-4">
          <div className={`${theme.dangerBg} rounded-xl p-4`}>
            <h4 className={`${theme.danger} font-semibold mb-3 flex items-center gap-2`}>
              <TrendingDown size={18} /> AVANT (Baseline)
            </h4>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className={theme.textMuted}>Durée</span>
                <span className={`font-mono ${theme.text}`}>{baselineMs.toFixed(0)} ms</span>
              </div>
              <div className="flex justify-between">
                <span className={theme.textMuted}>Coût annuel</span>
                <span className={`font-mono ${theme.danger}`}>{((baselineMs / 1000 / 3600) * costParams.executionsPerDay * costParams.workingDaysPerYear * costParams.hourlyRate).toFixed(0)} €</span>
              </div>
            </div>
          </div>

          <div className={`${theme.successBg} rounded-xl p-4`}>
            <h4 className={`${theme.success} font-semibold mb-3 flex items-center gap-2`}>
              <TrendingUp size={18} /> APRÈS (Optimisé)
            </h4>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className={theme.textMuted}>Durée</span>
                <span className={`font-mono ${theme.text}`}>{optimizedMs.toFixed(0)} ms</span>
              </div>
              <div className="flex justify-between">
                <span className={theme.textMuted}>Coût annuel</span>
                <span className={`font-mono ${theme.success}`}>{((optimizedMs / 1000 / 3600) * costParams.executionsPerDay * costParams.workingDaysPerYear * costParams.hourlyRate).toFixed(0)} €</span>
              </div>
            </div>
          </div>
        </div>

        {/* Explanation for non-technical people */}
        <div className={`mt-4 ${theme.bgTertiary} rounded-xl p-4`}>
          <p className={`text-sm ${theme.textSecondary}`}>
            <strong className={theme.text}>💡 En résumé :</strong> Chaque exécution est{' '}
            <span className={`font-semibold ${theme.success}`}>{data.comparison.gain_percent.toFixed(1)}% plus rapide</span>.
            Sur {costParams.executionsPerDay.toLocaleString()} exécutions par jour, pendant {costParams.workingDaysPerYear} jours par an,
            cela représente <span className="text-emerald-400 font-semibold">{annualTimeSavedHours.toFixed(1)} heures</span> de temps serveur économisées,
            soit <span className="text-emerald-400 font-semibold">{totalAnnualSavings.toFixed(0)}€</span> d'économies
            et <span className="text-green-400 font-semibold">{annualCo2SavedKg.toFixed(1)} kg de CO₂</span> en moins.
          </p>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid md:grid-cols-2 gap-4">
        <StatsCardThemed title="Statistiques Baseline" stats={data.baseline.stats} color="indigo" theme={theme} />
        <StatsCardThemed title="Statistiques Optimisé" stats={data.optimized.stats} color="emerald" theme={theme} />
      </div>

      {/* Duration Chart */}
      <div className={`${theme.card} rounded-xl p-6`}>
        <h3 className={`text-lg font-semibold ${theme.text} mb-4`}>Durées d'exécution</h3>
        <ResponsiveContainer width="100%" height={280}>
          <AreaChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke={theme.chartGrid} />
            <XAxis dataKey="name" stroke={theme.chartText} fontSize={12} />
            <YAxis stroke={theme.chartText} fontSize={12} />
            <Tooltip
              contentStyle={{ backgroundColor: theme.bgSecondary, border: `1px solid ${theme.chartGrid}`, borderRadius: '8px' }}
              labelStyle={{ color: theme.chartText }}
            />
            <Legend />
            <Area type="monotone" dataKey="baseline" stroke="#818cf8" fill="#818cf8" fillOpacity={0.1} strokeWidth={2} name="Baseline" />
            <Area type="monotone" dataKey="optimized" stroke="#34d399" fill="#34d399" fillOpacity={0.1} strokeWidth={2} name="Optimisé" />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* Config */}
      <div className={`${theme.card} rounded-xl p-6`}>
        <h3 className={`text-lg font-semibold ${theme.text} mb-4`}>Configuration du test</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
          <ConfigItemThemed label="Mode" value={data.config.mode} theme={theme} />
          <ConfigItemThemed label="Warmup" value={data.config.warmup_runs} theme={theme} />
          <ConfigItemThemed label="Runs" value={data.config.measured_runs} theme={theme} />
          <ConfigItemThemed label="Alternance" value={data.config.alternate ? 'Oui' : 'Non'} theme={theme} />
        </div>
      </div>
    </div>
  );
}

function AggregateReport({ data, costParams, theme }) {
  const gainPositive = data.aggregate_stats.median_gain >= 0;

  const chartData = data.reports.map(r => ({
    machine: r.machine,
    gain: r.comparison.gain_percent
  }));

  // Calculate aggregate financial impact across all machines
  const calculateMachineSavings = (report) => {
    const baselineMs = report.baseline.stats.median;
    const optimizedMs = report.optimized.stats.median;
    const timeSavedMs = baselineMs - optimizedMs;
    const timeSavedPerExec = timeSavedMs / 1000;
    const dailyTimeSavedSec = timeSavedPerExec * costParams.executionsPerDay;
    const dailyTimeSavedHours = dailyTimeSavedSec / 3600;
    const annualTimeSavedHours = dailyTimeSavedHours * costParams.workingDaysPerYear;
    const annualCostSavings = annualTimeSavedHours * costParams.hourlyRate;
    
    const baselineEnergyPerExec = (baselineMs / 1000 / 3600) * costParams.serverWatts;
    const optimizedEnergyPerExec = (optimizedMs / 1000 / 3600) * costParams.serverWatts;
    const energySavedPerExec = baselineEnergyPerExec - optimizedEnergyPerExec;
    const dailyEnergySavedWh = energySavedPerExec * costParams.executionsPerDay;
    const annualEnergySavedKwh = (dailyEnergySavedWh * costParams.workingDaysPerYear) / 1000;
    const annualElectricitySavings = annualEnergySavedKwh * costParams.electricityPrice;
    const annualCo2SavedKg = annualEnergySavedKwh * 0.4;
    
    return {
      annualCostSavings,
      annualElectricitySavings,
      annualEnergySavedKwh,
      annualCo2SavedKg,
      annualTimeSavedHours,
      total: annualCostSavings + annualElectricitySavings
    };
  };

  // Sum up savings across all machines
  const totalSavings = data.reports.reduce((acc, r) => {
    const savings = calculateMachineSavings(r);
    return {
      annualCostSavings: acc.annualCostSavings + savings.annualCostSavings,
      annualElectricitySavings: acc.annualElectricitySavings + savings.annualElectricitySavings,
      annualEnergySavedKwh: acc.annualEnergySavedKwh + savings.annualEnergySavedKwh,
      annualCo2SavedKg: acc.annualCo2SavedKg + savings.annualCo2SavedKg,
      annualTimeSavedHours: acc.annualTimeSavedHours + savings.annualTimeSavedHours,
      total: acc.total + savings.total
    };
  }, { annualCostSavings: 0, annualElectricitySavings: 0, annualEnergySavedKwh: 0, annualCo2SavedKg: 0, annualTimeSavedHours: 0, total: 0 });

  return (
    <div className="space-y-6">
      {/* Global Gain Card */}
      <div className={`${theme.card} rounded-2xl p-8 text-center`}>
        <div className="flex items-center justify-center gap-2 mb-4">
          <BarChart3 size={20} className={theme.accentText} />
          <h2 className={`text-lg font-semibold ${theme.textSecondary}`}>Gain Global Multi-Machines</h2>
        </div>
        <div className={`text-6xl font-bold mb-4 flex items-center justify-center gap-3 ${
          gainPositive ? theme.success : theme.danger
        }`}>
          {gainPositive ? <TrendingUp size={48} /> : <TrendingDown size={48} />}
          {data.aggregate_stats.median_gain.toFixed(2)}%
        </div>
        <p className={`mb-2 ${theme.textSecondary}`}>
          Gain médian sur <strong className={theme.text}>{data.machine_count}</strong> machines
        </p>
        <p className={`text-sm ${theme.textMuted}`}>
          P10/P90: {data.aggregate_stats.p10_gain.toFixed(2)}% / {data.aggregate_stats.p90_gain.toFixed(2)}%
        </p>
      </div>

      {/* Aggregate Financial Impact */}
      <div className={`${theme.card} rounded-2xl p-6 border-2 border-emerald-500/20`}>
        <div className="flex items-center gap-3 mb-6">
          <div className="p-3 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl shadow-lg shadow-emerald-500/25">
            <DollarSign size={24} className="text-white" />
          </div>
          <div>
            <h2 className={`text-xl font-bold ${theme.text}`}>Impact Financier Total ({data.machine_count} machines)</h2>
            <p className={`text-sm ${theme.textMuted}`}>Économies cumulées sur votre infrastructure</p>
          </div>
        </div>

        {/* Big Total Savings */}
        <div className="text-center mb-8">
          <p className={`text-sm mb-2 ${theme.textMuted}`}>Économies annuelles totales</p>
          <div className={`text-5xl font-bold ${totalSavings.total >= 0 ? theme.success : theme.danger}`}>
            {totalSavings.total >= 0 ? '+' : ''}{totalSavings.total.toLocaleString('fr-FR', { minimumFractionDigits: 0, maximumFractionDigits: 0 })} €
          </div>
          <p className={`text-sm mt-2 ${theme.textMuted}`}>par an</p>
        </div>

        {/* Impact Cards Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <ImpactCard icon={Clock} label="Temps économisé" value={`${totalSavings.annualTimeSavedHours.toFixed(0)}h`} subValue="par an" color="blue" theme={theme} />
          <ImpactCard icon={DollarSign} label="Coûts serveur" value={`+${totalSavings.annualCostSavings.toFixed(0)} €`} subValue="économisés /an" color="emerald" theme={theme} />
          <ImpactCard icon={Zap} label="Électricité" value={`${totalSavings.annualEnergySavedKwh.toFixed(0)} kWh`} subValue="économisés /an" color="amber" theme={theme} />
          <ImpactCard icon={Leaf} label="CO₂ évité" value={`-${totalSavings.annualCo2SavedKg.toFixed(0)} kg`} subValue="par an" color="green" theme={theme} />
        </div>

        {/* Summary for executives */}
        <div className={`mt-4 ${theme.bgTertiary} rounded-xl p-4`}>
          <p className={`text-sm ${theme.textSecondary}`}>
            <strong className={theme.text}>💼 Pour les décideurs :</strong> Sur vos{' '}
            <span className={`font-semibold ${theme.success}`}>{data.machine_count} machines</span>, économisez{' '}
            <span className={`font-semibold ${theme.success}`}>{totalSavings.total.toFixed(0)}€/an</span> et{' '}
            <span className="font-semibold text-green-500">{totalSavings.annualCo2SavedKg.toFixed(0)} kg CO₂</span>.
          </p>
        </div>
      </div>

      {/* Gains Chart */}
      <div className={`${theme.card} rounded-xl p-6`}>
        <h3 className={`text-lg font-semibold ${theme.text} mb-4`}>Distribution des gains par machine</h3>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke={theme.chartGrid} />
            <XAxis dataKey="machine" stroke={theme.chartText} fontSize={12} />
            <YAxis stroke={theme.chartText} fontSize={12} />
            <Tooltip
              contentStyle={{ backgroundColor: theme.bgSecondary, border: `1px solid ${theme.chartGrid}`, borderRadius: '8px' }}
              labelStyle={{ color: theme.chartText }}
            />
            <Bar
              dataKey="gain"
              fill="#34d399"
              radius={[4, 4, 0, 0]}
            />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Per-Machine Table */}
      <div className={`${theme.card} rounded-xl p-6`}>
        <h3 className={`text-lg font-semibold ${theme.text} mb-4`}>Résultats par machine</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className={`border-b ${theme.border}`}>
                <th className={`py-3 px-4 text-left ${theme.textMuted}`}>Machine</th>
                <th className={`py-3 px-4 text-right ${theme.textMuted}`}>Gain</th>
                <th className={`py-3 px-4 text-right ${theme.textMuted}`}>Baseline</th>
                <th className={`py-3 px-4 text-right ${theme.textMuted}`}>Optimisé</th>
                <th className={`py-3 px-4 text-center ${theme.textMuted}`}>Statut</th>
              </tr>
            </thead>
            <tbody>
              {data.reports.map((r, i) => (
                <tr key={i} className={`border-b ${theme.border} hover:${theme.bgTertiary}`}>
                  <td className={`py-3 px-4 font-medium ${theme.text}`}>{r.machine}</td>
                  <td className={`py-3 px-4 text-right font-mono ${
                    r.comparison.gain_percent >= 0 ? theme.success : theme.danger
                  }`}>
                    {r.comparison.gain_percent.toFixed(2)}%
                  </td>
                  <td className={`py-3 px-4 text-right font-mono ${theme.textSecondary}`}>
                    {r.baseline.stats.median.toFixed(0)} ms
                  </td>
                  <td className={`py-3 px-4 text-right font-mono ${theme.textSecondary}`}>
                    {r.optimized.stats.median.toFixed(0)} ms
                  </td>
                  <td className="py-3 px-4 text-center">
                    {r.comparison.conclusive ? (
                      <CheckCircle size={16} className={theme.success} />
                    ) : (
                      <AlertCircle size={16} className={theme.warning} />
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Note */}
      <div className={`${theme.bgTertiary} rounded-xl p-4`}>
        <p className={`text-sm ${theme.textSecondary}`}>
          <strong className={theme.text}>ℹ️ Note :</strong> Les comparaisons sont relatives (A vs B sur la même machine). 
          Les temps bruts ne sont jamais comparés entre machines.
        </p>
      </div>
    </div>
  );
}

function StatsCard({ title, stats, color }) {
  const colorClasses = {
    indigo: 'border-indigo-700 bg-indigo-900/20',
    emerald: 'border-emerald-700 bg-emerald-900/20'
  };

  return (
    <div className={`rounded-xl p-6 border ${colorClasses[color]}`}>
      <h3 className="text-lg font-semibold text-slate-300 mb-4">{title}</h3>
      <table className="w-full">
        <tbody className="text-sm">
          <StatRow label="Median" value={`${stats.median.toFixed(2)} ms`} />
          <StatRow label="Mean" value={`${stats.mean.toFixed(2)} ms`} />
          <StatRow label="Std Dev" value={`${stats.std_dev.toFixed(2)} ms`} />
          <StatRow label="CV" value={`${stats.cv.toFixed(2)}%`} />
          <StatRow label="P10" value={`${stats.p10.toFixed(2)} ms`} />
          <StatRow label="P90" value={`${stats.p90.toFixed(2)} ms`} />
        </tbody>
      </table>
    </div>
  );
}

function StatRow({ label, value }) {
  return (
    <tr className="border-b border-slate-700/50">
      <td className="py-2 text-slate-400">{label}</td>
      <td className="py-2 text-right font-mono text-slate-200">{value}</td>
    </tr>
  );
}

function ConfigItem({ label, value }) {
  return (
    <div>
      <span className="text-slate-500">{label}:</span>{' '}
      <code className="bg-slate-700 px-2 py-1 rounded text-slate-200">{value}</code>
    </div>
  );
}

function ConfigItemThemed({ label, value, theme }) {
  return (
    <div className={`${theme.bgTertiary} rounded-lg p-3`}>
      <div className={`text-xs ${theme.textMuted} mb-1`}>{label}</div>
      <div className={`font-medium ${theme.text}`}>{value}</div>
    </div>
  );
}

function CostInput({ label, value, onChange, theme }) {
  return (
    <div>
      <label className={`block text-sm mb-1.5 font-medium ${theme?.textSecondary || 'text-gray-600'}`}>
        {label}
      </label>
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
        className={`w-full rounded-lg px-3 py-2 text-sm border focus:outline-none focus:ring-2 focus:ring-indigo-500 ${theme?.input || 'bg-white border-gray-300 text-gray-900'}`}
      />
    </div>
  );
}

function ImpactCard({ icon: Icon, label, value, subValue, color, theme }) {
  const colors = {
    blue: 'text-blue-500',
    emerald: 'text-emerald-500',
    amber: 'text-amber-500',
    green: 'text-green-500',
  };

  return (
    <div className={`${theme.card} rounded-xl p-4`}>
      <div className="flex items-center gap-2 mb-2">
        <Icon size={18} className={colors[color]} />
        <span className={`text-xs ${theme.textMuted}`}>{label}</span>
      </div>
      <div className={`text-xl font-bold ${theme.text}`}>{value}</div>
      <div className={`text-xs mt-1 ${theme.textMuted}`}>{subValue}</div>
    </div>
  );
}

function StatsCardThemed({ title, stats, theme, color }) {
  const colors = {
    indigo: { border: 'border-indigo-500/30', bg: 'bg-indigo-500/5', dot: 'bg-indigo-500' },
    emerald: { border: 'border-emerald-500/30', bg: 'bg-emerald-500/5', dot: 'bg-emerald-500' }
  };
  const c = colors[color] || colors.indigo;

  return (
    <div className={`${theme.card} rounded-xl p-5 ${c.border} ${c.bg}`}>
      <div className="flex items-center gap-2 mb-4">
        <div className={`w-2 h-2 rounded-full ${c.dot}`}></div>
        <h3 className={`font-semibold ${theme.text}`}>{title}</h3>
      </div>
      <div className="space-y-2">
        {[
          { label: 'Médiane', value: `${stats.median.toFixed(2)} ms` },
          { label: 'Moyenne', value: `${stats.mean.toFixed(2)} ms` },
          { label: 'Écart-type', value: `${stats.std_dev.toFixed(2)} ms` },
          { label: 'CV', value: `${stats.cv.toFixed(2)}%` },
          { label: 'P10', value: `${stats.p10.toFixed(2)} ms` },
          { label: 'P90', value: `${stats.p90.toFixed(2)} ms` },
        ].map((row, i) => (
          <div key={i} className={`flex justify-between py-1.5 border-b ${theme.border} last:border-0`}>
            <span className={`text-sm ${theme.textMuted}`}>{row.label}</span>
            <span className={`text-sm font-mono ${theme.text}`}>{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
