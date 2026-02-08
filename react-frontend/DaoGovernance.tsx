// src/components/DaoGovernance.tsx - DAO Governance for 1000 University Founders
import React, { useState, useEffect } from 'react';
import { Users, Check, X, Globe, Award } from 'lucide-react';

interface Founder {
  id: string;
  name: string;
  country: string;
  address: string;
  verified: boolean;
}

interface Proposal {
  id: string;
  title: string;
  description: string;
  votes_for: number;
  votes_against: number;
  status: 'active' | 'passed' | 'rejected';
}

export const DaoGovernance: React.FC = () => {
  const [founders, setFounders] = useState<Founder[]>([]);
  const [selectedFounder, setSelectedFounder] = useState<string>('');
  const [proposals] = useState<Proposal[]>([
    {
      id: 'prop-1',
      title: 'Increase CXL Pool Capacity to 128GB',
      description: 'Expand the shared memory pool to accommodate growth in mesh nodes and enclave requirements.',
      votes_for: 647,
      votes_against: 123,
      status: 'active'
    },
    {
      id: 'prop-2',
      title: 'Implement Dynamic Stake Rewards',
      description: 'Adjust staking rewards based on network participation and contribution quality.',
      votes_for: 823,
      votes_against: 45,
      status: 'passed'
    },
    {
      id: 'prop-3',
      title: 'Enable Multi-Sig Treasury',
      description: 'Require 5-of-9 signatures for treasury operations over 1M tokens.',
      votes_for: 412,
      votes_against: 278,
      status: 'active'
    }
  ]);
  const [voteMessage, setVoteMessage] = useState('');

  // Fetch founders from backend
  useEffect(() => {
    const fetchFounders = async () => {
      try {
        const res = await fetch('http://localhost:5000/dao/founders');
        if (res.ok) {
          const json = await res.json();
          setFounders(json.founders);
        }
      } catch (error) {
        console.error('Failed to fetch founders:', error);
      }
    };

    fetchFounders();
  }, []);

  const handleVote = async (proposalId: string, vote: boolean) => {
    if (!selectedFounder) {
      setVoteMessage('Please select a founder to vote');
      return;
    }

    try {
      const res = await fetch('http://localhost:5000/dao/vote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          proposal_id: proposalId,
          voter_name: selectedFounder,
          vote
        })
      });

      if (res.ok) {
        const json = await res.json();
        if (json.success) {
          setVoteMessage(`✓ Vote recorded for ${selectedFounder}`);
          setTimeout(() => setVoteMessage(''), 3000);
        } else {
          setVoteMessage('✗ Vote failed - founder not verified');
        }
      }
    } catch (error) {
      console.error('Failed to submit vote:', error);
      setVoteMessage('✗ Network error');
    }
  };

  const calculateVotePercentage = (votesFor: number, votesAgainst: number) => {
    const total = votesFor + votesAgainst;
    return total > 0 ? (votesFor / total) * 100 : 0;
  };

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="bg-black/50 backdrop-blur-lg border border-purple-500/50 rounded-lg p-6">
          <h2 className="text-3xl font-bold text-purple-400 mb-2 flex items-center gap-3">
            <Users size={32} />
            DAO Governance
          </h2>
          <p className="text-gray-400">
            Decentralized governance by 1,000 university founders. 
            Each founder holds cryptographic signing rights for protocol decisions.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column: Founder Selection */}
          <div className="lg:col-span-1 space-y-6">
            <div className="bg-black/50 backdrop-blur-lg border border-purple-500/50 rounded-lg p-6">
              <h3 className="text-xl font-semibold text-purple-400 mb-4 flex items-center gap-2">
                <Award size={24} />
                Select Founder
              </h3>
              
              <select
                value={selectedFounder}
                onChange={(e) => setSelectedFounder(e.target.value)}
                className="w-full bg-gray-800 border border-purple-500/30 rounded-lg px-4 py-3 focus:outline-none focus:border-purple-500 text-white mb-4"
              >
                <option value="">Choose a university...</option>
                {founders.map((founder) => (
                  <option key={founder.id} value={founder.name}>
                    {founder.name} ({founder.country})
                  </option>
                ))}
              </select>

              {selectedFounder && (
                <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                  <p className="text-sm text-gray-400 mb-1">Selected:</p>
                  <p className="text-white font-semibold">{selectedFounder}</p>
                  {founders.find(f => f.name === selectedFounder)?.verified && (
                    <div className="mt-2 flex items-center gap-2 text-green-400">
                      <Check size={16} />
                      <span className="text-sm">Signature Verified</span>
                    </div>
                  )}
                </div>
              )}

              {voteMessage && (
                <div className={`mt-4 p-3 rounded-lg ${
                  voteMessage.includes('✓') 
                    ? 'bg-green-500/20 border border-green-500 text-green-400'
                    : 'bg-red-500/20 border border-red-500 text-red-400'
                }`}>
                  {voteMessage}
                </div>
              )}
            </div>

            {/* Founding Universities Stats */}
            <div className="bg-black/50 backdrop-blur-lg border border-purple-500/50 rounded-lg p-6">
              <h3 className="text-xl font-semibold text-purple-400 mb-4">Network Stats</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-400">Total Founders:</span>
                  <span className="text-2xl font-bold text-purple-400">1,000</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Verified:</span>
                  <span className="text-xl font-semibold text-green-400">
                    {founders.filter(f => f.verified).length}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Active Proposals:</span>
                  <span className="text-xl font-semibold text-blue-400">
                    {proposals.filter(p => p.status === 'active').length}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Proposals */}
          <div className="lg:col-span-2 space-y-4">
            <div className="bg-black/50 backdrop-blur-lg border border-purple-500/50 rounded-lg p-6">
              <h3 className="text-xl font-semibold text-purple-400 mb-4">Active Proposals</h3>
              
              <div className="space-y-4">
                {proposals.map((proposal) => {
                  const percentage = calculateVotePercentage(proposal.votes_for, proposal.votes_against);
                  const total = proposal.votes_for + proposal.votes_against;
                  
                  return (
                    <div
                      key={proposal.id}
                      className="bg-gray-800/50 border border-gray-700 rounded-lg p-5"
                    >
                      {/* Proposal Header */}
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex-1">
                          <h4 className="text-lg font-semibold text-white mb-2">
                            {proposal.title}
                          </h4>
                          <p className="text-sm text-gray-400">
                            {proposal.description}
                          </p>
                        </div>
                        <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                          proposal.status === 'passed' 
                            ? 'bg-green-500/20 text-green-400'
                            : proposal.status === 'rejected'
                            ? 'bg-red-500/20 text-red-400'
                            : 'bg-blue-500/20 text-blue-400'
                        }`}>
                          {proposal.status.toUpperCase()}
                        </span>
                      </div>

                      {/* Vote Progress */}
                      <div className="mb-4">
                        <div className="flex justify-between text-sm mb-2">
                          <span className="text-green-400">For: {proposal.votes_for}</span>
                          <span className="text-red-400">Against: {proposal.votes_against}</span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-3 overflow-hidden">
                          <div 
                            className="bg-gradient-to-r from-green-500 to-green-600 h-3 transition-all"
                            style={{ width: `${percentage}%` }}
                          />
                        </div>
                        <div className="flex justify-between text-xs text-gray-500 mt-1">
                          <span>{percentage.toFixed(1)}% approval</span>
                          <span>{total} total votes</span>
                        </div>
                      </div>

                      {/* Vote Buttons */}
                      {proposal.status === 'active' && (
                        <div className="flex gap-3">
                          <button
                            onClick={() => handleVote(proposal.id, true)}
                            disabled={!selectedFounder}
                            className="flex-1 bg-green-500 hover:bg-green-600 disabled:bg-gray-700 disabled:text-gray-500 px-4 py-2 rounded-lg flex items-center justify-center gap-2 transition"
                          >
                            <Check size={20} />
                            Vote For
                          </button>
                          <button
                            onClick={() => handleVote(proposal.id, false)}
                            disabled={!selectedFounder}
                            className="flex-1 bg-red-500 hover:bg-red-600 disabled:bg-gray-700 disabled:text-gray-500 px-4 py-2 rounded-lg flex items-center justify-center gap-2 transition"
                          >
                            <X size={20} />
                            Vote Against
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>

        {/* Founding Universities Map */}
        <div className="bg-black/50 backdrop-blur-lg border border-purple-500/50 rounded-lg p-6">
          <h3 className="text-xl font-semibold text-purple-400 mb-4 flex items-center gap-2">
            <Globe size={24} />
            Global University Network (First 20 Shown)
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
            {founders.map((founder) => (
              <div
                key={founder.id}
                className="bg-gray-800/50 border border-gray-700 rounded-lg p-3 hover:border-purple-500/50 transition cursor-pointer"
                onClick={() => setSelectedFounder(founder.name)}
              >
                <div className="flex items-start justify-between mb-1">
                  <span className="text-sm font-semibold text-white">{founder.name}</span>
                  {founder.verified && (
                    <Check size={14} className="text-green-400" />
                  )}
                </div>
                <p className="text-xs text-gray-400">{founder.country}</p>
                <p className="text-xs text-gray-500 font-mono mt-1">{founder.address}</p>
              </div>
            ))}
          </div>
          
          <p className="text-sm text-gray-500 mt-4 text-center">
            + 980 more universities worldwide
          </p>
        </div>
      </div>
    </div>
  );
};
