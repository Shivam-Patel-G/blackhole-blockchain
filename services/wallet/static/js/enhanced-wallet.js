// Enhanced BlackHole Wallet JavaScript
class BlackHoleWallet {
    constructor() {
        this.currentSection = 'overview';
        this.wallets = [];
        this.balances = {};
        this.transactions = [];
        this.websocket = null;
        this.updateInterval = null;
        
        this.init();
    }

    init() {
        this.loadInitialData();
        this.setupWebSocket();
        this.startRealTimeUpdates();
        this.showSection('overview');
    }

    // Sidebar Management
    toggleSidebar() {
        const sidebar = document.getElementById('sidebar');
        const mainContent = document.getElementById('mainContent');
        const toggleIcon = document.getElementById('toggleIcon');
        
        sidebar.classList.toggle('collapsed');
        mainContent.classList.toggle('expanded');
        
        if (sidebar.classList.contains('collapsed')) {
            toggleIcon.className = 'fas fa-chevron-right';
        } else {
            toggleIcon.className = 'fas fa-chevron-left';
        }
    }

    // Section Management
    showSection(sectionName) {
        // Update active nav item
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        event.target.closest('.nav-item').classList.add('active');
        
        // Update page title
        const titles = {
            'overview': 'Dashboard Overview',
            'portfolio': 'Portfolio Analysis',
            'wallets': 'My Wallets',
            'transfer': 'Send Tokens',
            'receive': 'Receive Tokens',
            'history': 'Transaction History',
            'staking': 'Staking Dashboard',
            'dex': 'DEX Trading',
            'otc': 'OTC Trading',
            'bridge': 'Cross-Chain Bridge',
            'proposals': 'Governance Proposals',
            'voting': 'Voting Dashboard',
            'security': 'Security Settings',
            'settings': 'Preferences'
        };
        
        document.getElementById('pageTitle').textContent = titles[sectionName] || 'Dashboard';
        this.currentSection = sectionName;
        
        // Load section content
        this.loadSectionContent(sectionName);
    }

    // Content Loading
    loadSectionContent(section) {
        const contentArea = document.getElementById('contentArea');
        
        switch(section) {
            case 'overview':
                contentArea.innerHTML = this.getOverviewContent();
                this.loadOverviewData();
                break;
            case 'portfolio':
                contentArea.innerHTML = this.getPortfolioContent();
                this.loadPortfolioData();
                break;
            case 'wallets':
                contentArea.innerHTML = this.getWalletsContent();
                this.loadWalletsData();
                break;
            case 'transfer':
                contentArea.innerHTML = this.getTransferContent();
                break;
            case 'receive':
                contentArea.innerHTML = this.getReceiveContent();
                break;
            case 'history':
                contentArea.innerHTML = this.getHistoryContent();
                this.loadTransactionHistory();
                break;
            case 'staking':
                contentArea.innerHTML = this.getStakingContent();
                this.loadStakingData();
                break;
            case 'dex':
                contentArea.innerHTML = this.getDEXContent();
                break;
            case 'otc':
                contentArea.innerHTML = this.getOTCContent();
                break;
            case 'bridge':
                contentArea.innerHTML = this.getBridgeContent();
                break;
            case 'proposals':
                contentArea.innerHTML = this.getProposalsContent();
                this.loadGovernanceData();
                break;
            case 'voting':
                contentArea.innerHTML = this.getVotingContent();
                break;
            case 'security':
                contentArea.innerHTML = this.getSecurityContent();
                break;
            case 'settings':
                contentArea.innerHTML = this.getSettingsContent();
                break;
            default:
                contentArea.innerHTML = this.getOverviewContent();
        }
    }

    // Overview Content
    getOverviewContent() {
        return `
            <div class="grid grid-4">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-wallet"></i>
                            Total Balance
                        </div>
                    </div>
                    <div style="font-size: 32px; font-weight: bold; color: #667eea;" id="totalBalance">
                        <div class="loading"></div>
                    </div>
                    <div style="color: #666; margin-top: 5px;">Across all wallets</div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-coins"></i>
                            Staked Amount
                        </div>
                    </div>
                    <div style="font-size: 32px; font-weight: bold; color: #28a745;" id="stakedAmount">
                        <div class="loading"></div>
                    </div>
                    <div style="color: #666; margin-top: 5px;">Currently staking</div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-chart-line"></i>
                            Pending Rewards
                        </div>
                    </div>
                    <div style="font-size: 32px; font-weight: bold; color: #ffc107;" id="pendingRewards">
                        <div class="loading"></div>
                    </div>
                    <div style="color: #666; margin-top: 5px;">Ready to claim</div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-network-wired"></i>
                            Network Status
                        </div>
                    </div>
                    <div id="networkStatus">
                        <div class="status-indicator status-online">
                            <i class="fas fa-circle"></i>
                            Connected
                        </div>
                        <div style="margin-top: 10px; font-size: 14px; color: #666;">
                            Block Height: <span id="blockHeight">Loading...</span>
                        </div>
                    </div>
                </div>
            </div>

            <div class="grid grid-2" style="margin-top: 25px;">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-clock"></i>
                            Recent Transactions
                        </div>
                        <button class="btn btn-secondary" onclick="wallet.showSection('history')">
                            View All
                        </button>
                    </div>
                    <div id="recentTransactions">
                        <div class="loading"></div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-chart-pie"></i>
                            Token Distribution
                        </div>
                    </div>
                    <div id="tokenDistribution">
                        <div class="loading"></div>
                    </div>
                </div>
            </div>

            <div class="card" style="margin-top: 25px;">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-bolt"></i>
                        Quick Actions
                    </div>
                </div>
                <div class="grid grid-4">
                    <button class="btn btn-success" onclick="wallet.showSection('transfer')">
                        <i class="fas fa-paper-plane"></i>
                        Send Tokens
                    </button>
                    <button class="btn btn-info" onclick="wallet.showSection('receive')">
                        <i class="fas fa-qrcode"></i>
                        Receive Tokens
                    </button>
                    <button class="btn btn-warning" onclick="wallet.showSection('staking')">
                        <i class="fas fa-coins"></i>
                        Stake Tokens
                    </button>
                    <button class="btn" onclick="wallet.showSection('dex')">
                        <i class="fas fa-exchange-alt"></i>
                        Trade on DEX
                    </button>
                </div>
            </div>
        `;
    }

    // Wallets Content
    getWalletsContent() {
        return `
            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-wallet"></i>
                        Wallet Management
                    </div>
                    <button class="btn btn-success" onclick="wallet.showCreateWalletModal()">
                        <i class="fas fa-plus"></i>
                        Create Wallet
                    </button>
                </div>
                <div id="walletsList">
                    <div class="loading"></div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-coins"></i>
                        Token Balances
                    </div>
                    <button class="btn btn-secondary" onclick="wallet.refreshBalances()">
                        <i class="fas fa-sync"></i>
                        Refresh
                    </button>
                </div>
                <div id="balancesList">
                    <div class="loading"></div>
                </div>
            </div>
        `;
    }

    // Transfer Content
    getTransferContent() {
        return `
            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-paper-plane"></i>
                        Send Tokens
                    </div>
                </div>
                <form id="transferForm" onsubmit="wallet.submitTransfer(event)">
                    <div class="grid grid-2">
                        <div>
                            <label>From Wallet</label>
                            <select id="fromWallet" required>
                                <option value="">Select wallet...</option>
                            </select>
                        </div>
                        <div>
                            <label>Token</label>
                            <select id="tokenSelect" required>
                                <option value="BHX">BHX</option>
                                <option value="ETH">ETH</option>
                                <option value="USDT">USDT</option>
                            </select>
                        </div>
                    </div>
                    <div>
                        <label>Recipient Address</label>
                        <input type="text" id="recipientAddress" placeholder="Enter recipient address" required>
                    </div>
                    <div>
                        <label>Amount</label>
                        <input type="number" id="transferAmount" placeholder="0.00" step="0.000001" required>
                        <div style="margin-top: 5px; font-size: 12px; color: #666;">
                            Available: <span id="availableBalance">0</span> <span id="selectedToken">BHX</span>
                        </div>
                    </div>
                    <div>
                        <label>Gas Fee (optional)</label>
                        <input type="number" id="gasFee" placeholder="Auto" step="0.000001">
                    </div>
                    <button type="submit" class="btn btn-success" style="width: 100%; margin-top: 20px;">
                        <i class="fas fa-paper-plane"></i>
                        Send Transaction
                    </button>
                </form>
                <div id="transferResult" style="margin-top: 20px;"></div>
            </div>

            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-history"></i>
                        Recent Transfers
                    </div>
                </div>
                <div id="recentTransfers">
                    <div class="loading"></div>
                </div>
            </div>
        `;
    }

    // API Methods
    async loadInitialData() {
        try {
            await Promise.all([
                this.loadWallets(),
                this.loadBalances(),
                this.loadNetworkStatus()
            ]);
        } catch (error) {
            console.error('Error loading initial data:', error);
        }
    }

    async loadWallets() {
        try {
            const response = await fetch('/api/wallets');
            const data = await response.json();
            if (data.success) {
                this.wallets = data.wallets || [];
            }
        } catch (error) {
            console.error('Error loading wallets:', error);
        }
    }

    async loadBalances() {
        try {
            const response = await fetch('/api/balances');
            const data = await response.json();
            if (data.success) {
                this.balances = data.balances || {};
            }
        } catch (error) {
            console.error('Error loading balances:', error);
        }
    }

    async loadNetworkStatus() {
        try {
            const response = await fetch('http://localhost:8080/api/status');
            const data = await response.json();
            if (data.success) {
                this.updateNetworkStatus(data.data);
            }
        } catch (error) {
            console.error('Error loading network status:', error);
        }
    }

    // Real-time Updates
    setupWebSocket() {
        // WebSocket connection for real-time updates
        try {
            this.websocket = new WebSocket('ws://localhost:9000/ws');
            
            this.websocket.onopen = () => {
                console.log('WebSocket connected');
            };
            
            this.websocket.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.handleRealTimeUpdate(data);
            };
            
            this.websocket.onclose = () => {
                console.log('WebSocket disconnected');
                // Attempt to reconnect after 5 seconds
                setTimeout(() => this.setupWebSocket(), 5000);
            };
        } catch (error) {
            console.error('WebSocket connection failed:', error);
        }
    }

    startRealTimeUpdates() {
        // Update data every 30 seconds
        this.updateInterval = setInterval(() => {
            this.loadNetworkStatus();
            if (this.currentSection === 'overview') {
                this.loadOverviewData();
            }
        }, 30000);
    }

    handleRealTimeUpdate(data) {
        switch (data.type) {
            case 'balance_update':
                this.updateBalance(data.address, data.token, data.balance);
                break;
            case 'new_transaction':
                this.addNewTransaction(data.transaction);
                break;
            case 'block_update':
                this.updateBlockHeight(data.height);
                break;
        }
    }

    // Additional Content Methods
    getStakingContent() {
        return `
            <div class="grid grid-2">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-coins"></i>
                            Stake Tokens
                        </div>
                    </div>
                    <form id="stakingForm" onsubmit="wallet.submitStaking(event)">
                        <div>
                            <label>Validator</label>
                            <select id="validatorSelect" required>
                                <option value="">Select validator...</option>
                                <option value="genesis-validator">Genesis Validator</option>
                            </select>
                        </div>
                        <div>
                            <label>Amount to Stake</label>
                            <input type="number" id="stakeAmount" placeholder="0.00" step="0.000001" required>
                        </div>
                        <button type="submit" class="btn btn-warning" style="width: 100%; margin-top: 15px;">
                            <i class="fas fa-coins"></i>
                            Stake Tokens
                        </button>
                    </form>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-chart-line"></i>
                            Staking Overview
                        </div>
                    </div>
                    <div id="stakingOverview">
                        <div class="loading"></div>
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-users"></i>
                        Active Validators
                    </div>
                </div>
                <div id="validatorsList">
                    <div class="loading"></div>
                </div>
            </div>
        `;
    }

    getProposalsContent() {
        return `
            <div class="card">
                <div class="card-header">
                    <div class="card-title">
                        <i class="fas fa-vote-yea"></i>
                        Governance Proposals
                    </div>
                    <button class="btn btn-success" onclick="wallet.showCreateProposalModal()">
                        <i class="fas fa-plus"></i>
                        Create Proposal
                    </button>
                </div>
                <div id="proposalsList">
                    <div class="loading"></div>
                </div>
            </div>
        `;
    }

    getDEXContent() {
        return `
            <div class="grid grid-2">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-exchange-alt"></i>
                            Token Swap
                        </div>
                    </div>
                    <form id="swapForm" onsubmit="wallet.submitSwap(event)">
                        <div>
                            <label>From Token</label>
                            <select id="fromToken" required>
                                <option value="BHX">BHX</option>
                                <option value="ETH">ETH</option>
                                <option value="USDT">USDT</option>
                            </select>
                        </div>
                        <div>
                            <label>To Token</label>
                            <select id="toToken" required>
                                <option value="ETH">ETH</option>
                                <option value="BHX">BHX</option>
                                <option value="USDT">USDT</option>
                            </select>
                        </div>
                        <div>
                            <label>Amount</label>
                            <input type="number" id="swapAmount" placeholder="0.00" step="0.000001" required>
                        </div>
                        <button type="submit" class="btn btn-info" style="width: 100%; margin-top: 15px;">
                            <i class="fas fa-exchange-alt"></i>
                            Swap Tokens
                        </button>
                    </form>
                </div>

                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            <i class="fas fa-chart-area"></i>
                            Liquidity Pools
                        </div>
                    </div>
                    <div id="liquidityPools">
                        <div class="loading"></div>
                    </div>
                </div>
            </div>
        `;
    }

    // Data Loading Methods
    async loadOverviewData() {
        try {
            // Load total balance
            const balanceResponse = await fetch('/api/total-balance');
            const balanceData = await balanceResponse.json();
            if (balanceData.success) {
                document.getElementById('totalBalance').textContent =
                    this.formatBalance(balanceData.total) + ' BHX';
            }

            // Load staking data
            const stakingResponse = await fetch('/api/staking-overview');
            const stakingData = await stakingResponse.json();
            if (stakingData.success) {
                document.getElementById('stakedAmount').textContent =
                    this.formatBalance(stakingData.staked) + ' BHX';
                document.getElementById('pendingRewards').textContent =
                    this.formatBalance(stakingData.rewards) + ' BHX';
            }
        } catch (error) {
            console.error('Error loading overview data:', error);
        }
    }

    async loadGovernanceData() {
        try {
            const response = await fetch('http://localhost:8080/api/governance/proposals');
            const data = await response.json();
            if (data.success) {
                this.displayProposals(data.proposals);
            }
        } catch (error) {
            console.error('Error loading governance data:', error);
        }
    }

    displayProposals(proposals) {
        const proposalsList = document.getElementById('proposalsList');
        if (!proposals || proposals.length === 0) {
            proposalsList.innerHTML = '<p>No proposals found.</p>';
            return;
        }

        proposalsList.innerHTML = proposals.map(proposal => `
            <div class="card" style="margin-bottom: 15px;">
                <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 10px;">
                    <h4>${proposal.title}</h4>
                    <span class="status-indicator ${proposal.status === 'active' ? 'status-online' : 'status-pending'}">
                        ${proposal.status}
                    </span>
                </div>
                <p style="color: #666; margin-bottom: 15px;">${proposal.description}</p>
                <div style="display: flex; gap: 10px;">
                    <button class="btn btn-success" onclick="wallet.voteOnProposal('${proposal.id}', 'yes')">
                        <i class="fas fa-thumbs-up"></i> Yes (${proposal.votes.yes})
                    </button>
                    <button class="btn btn-danger" onclick="wallet.voteOnProposal('${proposal.id}', 'no')">
                        <i class="fas fa-thumbs-down"></i> No (${proposal.votes.no})
                    </button>
                    <button class="btn btn-secondary" onclick="wallet.voteOnProposal('${proposal.id}', 'abstain')">
                        <i class="fas fa-minus"></i> Abstain (${proposal.votes.abstain})
                    </button>
                </div>
            </div>
        `).join('');
    }

    // Action Methods
    async submitTransfer(event) {
        event.preventDefault();
        const formData = new FormData(event.target);

        try {
            const response = await fetch('http://localhost:8080/api/token/transfer', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    from: document.getElementById('fromWallet').value,
                    to: document.getElementById('recipientAddress').value,
                    amount: parseFloat(document.getElementById('transferAmount').value),
                    token: document.getElementById('tokenSelect').value
                })
            });

            const result = await response.json();
            this.displayTransferResult(result);
        } catch (error) {
            console.error('Transfer error:', error);
        }
    }

    async voteOnProposal(proposalId, option) {
        try {
            const response = await fetch('http://localhost:8080/api/governance/proposal/vote', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    proposal_id: proposalId,
                    voter: 'current-user',
                    option: option
                })
            });

            const result = await response.json();
            if (result.success) {
                alert('Vote cast successfully!');
                this.loadGovernanceData(); // Refresh proposals
            }
        } catch (error) {
            console.error('Voting error:', error);
        }
    }

    // Utility Methods
    formatBalance(balance) {
        return parseFloat(balance).toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 6
        });
    }

    formatAddress(address) {
        if (!address) return '';
        return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
    }

    updateNetworkStatus(data) {
        const blockHeight = document.getElementById('blockHeight');
        if (blockHeight) {
            blockHeight.textContent = data.block_height;
        }
    }

    async logout() {
        try {
            await fetch('/api/logout', { method: 'POST' });
            window.location.href = '/login';
        } catch (error) {
            console.error('Logout error:', error);
            window.location.href = '/login';
        }
    }
}

// Global wallet instance
let wallet;

// Initialize when page loads
document.addEventListener('DOMContentLoaded', () => {
    wallet = new BlackHoleWallet();
});

// Global functions for HTML onclick events
function toggleSidebar() {
    wallet.toggleSidebar();
}

function showSection(section) {
    wallet.showSection(section);
}

function logout() {
    wallet.logout();
}
