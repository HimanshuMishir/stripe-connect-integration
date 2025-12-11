import { useState, useEffect } from 'react'
import axios from 'axios'

function DeveloperDashboard({ orgId, jwtToken }) {
  const [status, setStatus] = useState(null)
  const [balance, setBalance] = useState(null)
  const [transactions, setTransactions] = useState([])
  const [withdrawals, setWithdrawals] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [success, setSuccess] = useState(null)
  const [withdrawAmount, setWithdrawAmount] = useState(50)

  const API_URL = 'http://localhost:8080/api/v1/connect'

  useEffect(() => {
    if (jwtToken) {
      loadDashboardData()
    }
  }, [orgId, jwtToken])

  const getHeaders = () => ({
    'X-Organization-ID': orgId,
    'Authorization': `Bearer ${jwtToken}`,
  })

  const loadDashboardData = async () => {
    if (!jwtToken) {
      setError('JWT token is required. Please add your token in the form above.')
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const headers = getHeaders()

      const [statusRes, balanceRes, transactionsRes, withdrawalsRes] = await Promise.allSettled([
        axios.get(`${API_URL}/status`, { headers }),
        axios.get(`${API_URL}/wallet/balance`, { headers }),
        axios.get(`${API_URL}/wallet/transactions?limit=10`, { headers }),
        axios.get(`${API_URL}/withdrawals/history?limit=10`, { headers }),
      ])

      if (statusRes.status === 'fulfilled') setStatus(statusRes.value.data)
      if (balanceRes.status === 'fulfilled') setBalance(balanceRes.value.data)
      if (transactionsRes.status === 'fulfilled') setTransactions(transactionsRes.value.data.transactions || [])
      if (withdrawalsRes.status === 'fulfilled') setWithdrawals(withdrawalsRes.value.data.withdrawals || [])
    } catch (err) {
      console.error('Error loading dashboard:', err)
      setError(err.response?.data?.error || 'Failed to load dashboard data')
    } finally {
      setLoading(false)
    }
  }

  const handleOnboard = async () => {
    if (!jwtToken) {
      setError('JWT token is required')
      return
    }

    setError(null)
    setSuccess(null)

    try {
      const headers = getHeaders()
      const response = await axios.post(`${API_URL}/onboard`, {
        refresh_url: window.location.origin + '/connect/refresh',
        return_url: window.location.origin + '/connect/complete',
      }, { headers })

      const { onboarding_url } = response.data
      window.open(onboarding_url, '_blank')
      setSuccess('Onboarding link opened in new tab. Complete the process and refresh this page.')
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create onboarding link')
    }
  }

  const handleWithdraw = async () => {
    if (!jwtToken) {
      setError('JWT token is required')
      return
    }

    setError(null)
    setSuccess(null)

    if (withdrawAmount < 50) {
      setError('Minimum withdrawal amount is $50')
      return
    }

    if (!balance || balance.balance < withdrawAmount) {
      setError('Insufficient balance')
      return
    }

    try {
      const headers = getHeaders()
      const response = await axios.post(`${API_URL}/withdrawals/request`, {
        amount: withdrawAmount,
      }, { headers })

      setSuccess(`Withdrawal request created! ${response.data.message}`)
      loadDashboardData()
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to request withdrawal')
    }
  }

  if (loading) {
    return <div className="loading">Loading dashboard...</div>
  }

  const isOnboarded = status?.onboarding_completed

  return (
    <div>
      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      {/* Onboarding Status */}
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h2 style={{ margin: 0 }}>Stripe Connect Status</h2>
          <button
            className="btn btn-secondary"
            onClick={loadDashboardData}
            disabled={loading}
            style={{ fontSize: '14px', padding: '8px 16px' }}
          >
            🔄 Refresh Status
          </button>
        </div>
        {!isOnboarded ? (
          <div>
            <p style={{ marginBottom: '16px', color: '#666' }}>
              Complete Stripe Connect onboarding to start receiving payments from users.
            </p>
            <button className="btn btn-primary" onClick={handleOnboard}>
              Start Onboarding
            </button>
            <p style={{ marginTop: '12px', fontSize: '13px', color: '#f59e0b' }}>
              💡 <strong>Tip:</strong> After completing onboarding, click "Refresh Status" to update your account status.
            </p>
          </div>
        ) : (
          <div>
            <p style={{ marginBottom: '12px' }}>
              <span className="status-badge success">Onboarding Complete</span>
            </p>
            <div className="info-grid">
              <div className="info-item">
                <label>Account ID</label>
                <div className="value" style={{ fontSize: '14px' }}>
                  {status.account_id || 'N/A'}
                </div>
              </div>
              <div className="info-item">
                <label>Payouts Enabled</label>
                <div className="value" style={{ fontSize: '16px' }}>
                  {status.payouts_enabled ? '✅ Yes' : '❌ No'}
                </div>
              </div>
              <div className="info-item">
                <label>Charges Enabled</label>
                <div className="value" style={{ fontSize: '16px' }}>
                  {status.charges_enabled ? '✅ Yes' : '❌ No'}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Wallet Balance */}
      {isOnboarded && balance && (
        <div className="card">
          <h2>Wallet Balance</h2>
          <div className="info-grid">
            <div className="info-item">
              <label>Current Balance</label>
              <div className="value success">${balance.balance.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Total Earned</label>
              <div className="value">${balance.total_earned.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Total Withdrawn</label>
              <div className="value">${balance.total_withdrawn.toFixed(2)}</div>
            </div>
            <div className="info-item">
              <label>Pending Withdrawals</label>
              <div className="value warning">${balance.pending_withdrawals.toFixed(2)}</div>
            </div>
          </div>

          {balance.can_withdraw && (
            <div style={{ marginTop: '20px' }}>
              <h3>Request Withdrawal</h3>
              <p style={{ color: '#666', marginBottom: '12px', fontSize: '14px' }}>
                Minimum withdrawal: ${balance.minimum_withdrawal.toFixed(2)}
              </p>
              <div style={{ display: 'flex', gap: '10px', alignItems: 'flex-end' }}>
                <div className="form-group" style={{ flex: 1, maxWidth: '200px', marginBottom: 0 }}>
                  <label htmlFor="withdrawAmount">Amount (USD)</label>
                  <input
                    id="withdrawAmount"
                    type="number"
                    min="50"
                    step="0.01"
                    value={withdrawAmount}
                    onChange={(e) => setWithdrawAmount(parseFloat(e.target.value))}
                  />
                </div>
                <button
                  className="btn btn-primary"
                  onClick={handleWithdraw}
                  disabled={withdrawAmount < 50 || withdrawAmount > balance.balance}
                >
                  Withdraw ${withdrawAmount.toFixed(2)}
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Transaction History */}
      {isOnboarded && transactions.length > 0 && (
        <div className="card">
          <h2>Recent Transactions</h2>
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Function</th>
                <th>User Org</th>
                <th>Amount</th>
                <th>Platform Fee</th>
                <th>Net Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => (
                <tr key={tx.id}>
                  <td>{new Date(tx.executed_at).toLocaleDateString()}</td>
                  <td>{tx.function_name}</td>
                  <td style={{ fontSize: '12px' }}>{tx.user_organization.substring(0, 12)}...</td>
                  <td>${tx.amount.toFixed(2)}</td>
                  <td>${tx.platform_fee.toFixed(2)}</td>
                  <td style={{ fontWeight: 'bold' }}>${tx.net_amount.toFixed(2)}</td>
                  <td>
                    <span className={`status-badge ${tx.status === 'completed' ? 'success' : 'pending'}`}>
                      {tx.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Withdrawal History */}
      {isOnboarded && withdrawals.length > 0 && (
        <div className="card">
          <h2>Withdrawal History</h2>
          <table>
            <thead>
              <tr>
                <th>Requested</th>
                <th>Amount</th>
                <th>Status</th>
                <th>Completed</th>
                <th>Reason</th>
              </tr>
            </thead>
            <tbody>
              {withdrawals.map((wd) => (
                <tr key={wd.id}>
                  <td>{new Date(wd.requested_at).toLocaleDateString()}</td>
                  <td>${wd.amount.toFixed(2)}</td>
                  <td>
                    <span className={`status-badge ${
                      wd.status === 'completed' ? 'success' :
                      wd.status === 'failed' ? 'failed' :
                      'pending'
                    }`}>
                      {wd.status}
                    </span>
                  </td>
                  <td>{wd.completed_at ? new Date(wd.completed_at).toLocaleDateString() : '-'}</td>
                  <td>{wd.failure_reason || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

export default DeveloperDashboard
