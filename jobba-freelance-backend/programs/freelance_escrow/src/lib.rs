use anchor_lang::prelude::*;
use anchor_spl::{
    associated_token::AssociatedToken,
    token_interface::{self, CloseAccount, Mint, TokenAccount, TokenInterface, TransferChecked},
};

declare_id!("CwNiWw9TSRDnszphzCHsShxC64VhgbawWxCLnZ9gLhxy");

const BPS_DENOMINATOR: u64 = 10_000;
const CURRENT_VERSION: u8 = 1;

#[program]
pub mod freelance_escrow {
    use super::*;

    pub fn initialize_config(ctx: Context<InitializeConfig>, fee_bps: u16) -> Result<()> {
        require!(fee_bps as u64 <= BPS_DENOMINATOR, EscrowError::InvalidFeeBps);
        require!(ctx.accounts.treasury.key() != Pubkey::default(), EscrowError::InvalidTreasury);
        let config = &mut ctx.accounts.config;
        config.version = CURRENT_VERSION;
        config.admin = ctx.accounts.admin.key();
        config.pending_admin = Pubkey::default();
        config.accepted_mint = ctx.accounts.accepted_mint.key();
        config.token_program = ctx.accounts.token_program.key();
        config.treasury = ctx.accounts.treasury.key();
        config.fee_bps = fee_bps;
        config.paused = false;
        config.bump = ctx.bumps.config;
        emit!(ConfigInitialized {
            config: config.key(),
            admin: config.admin,
            mint: config.accepted_mint,
            treasury: config.treasury,
            fee_bps,
        });
        Ok(())
    }

    pub fn propose_admin(ctx: Context<AdminConfig>, pending_admin: Pubkey) -> Result<()> {
        require!(pending_admin != Pubkey::default(), EscrowError::InvalidAdmin);
        ctx.accounts.config.pending_admin = pending_admin;
        emit!(AdminProposed { config: ctx.accounts.config.key(), pending_admin });
        Ok(())
    }

    pub fn accept_admin(ctx: Context<AcceptAdmin>) -> Result<()> {
        let config = &mut ctx.accounts.config;
        require_keys_eq!(config.pending_admin, ctx.accounts.pending_admin.key(), EscrowError::Unauthorized);
        let previous_admin = config.admin;
        config.admin = config.pending_admin;
        config.pending_admin = Pubkey::default();
        emit!(AdminAccepted { config: config.key(), previous_admin, admin: config.admin });
        Ok(())
    }

    pub fn set_paused(ctx: Context<AdminConfig>, paused: bool) -> Result<()> {
        ctx.accounts.config.paused = paused;
        emit!(PauseChanged { config: ctx.accounts.config.key(), paused });
        Ok(())
    }

    pub fn create_and_fund_escrow(
        ctx: Context<CreateAndFundEscrow>,
        contract_id_hash: [u8; 32],
        terms_hash: [u8; 32],
        gross_amount: u64,
    ) -> Result<()> {
        require!(!ctx.accounts.config.paused, EscrowError::Paused);
        require!(gross_amount > 0, EscrowError::InvalidAmount);
        require!(terms_hash != [0; 32], EscrowError::InvalidTermsHash);
        require!(contract_id_hash != [0; 32], EscrowError::InvalidContractHash);
        require_keys_neq!(ctx.accounts.client.key(), ctx.accounts.freelancer.key(), EscrowError::SameParty);

        let fee_amount = calculate_fee(gross_amount, ctx.accounts.config.fee_bps)?;
        transfer_from_user(
            &ctx.accounts.token_program,
            &ctx.accounts.client_token,
            &ctx.accounts.vault,
            &ctx.accounts.accepted_mint,
            &ctx.accounts.client,
            gross_amount,
        )?;

        let now = Clock::get()?.unix_timestamp;
        let escrow = &mut ctx.accounts.escrow;
        escrow.version = CURRENT_VERSION;
        escrow.config = ctx.accounts.config.key();
        escrow.contract_id_hash = contract_id_hash;
        escrow.terms_hash = terms_hash;
        escrow.client = ctx.accounts.client.key();
        escrow.freelancer = ctx.accounts.freelancer.key();
        escrow.mint = ctx.accounts.accepted_mint.key();
        escrow.gross_amount = gross_amount;
        escrow.fee_amount = fee_amount;
        escrow.state = EscrowState::Funded;
        escrow.created_at = now;
        escrow.accepted_at = 0;
        escrow.submitted_at = 0;
        escrow.disputed_at = 0;
        escrow.resolved_at = 0;
        escrow.bump = ctx.bumps.escrow;

        emit!(EscrowFunded {
            escrow: escrow.key(), contract_id_hash, terms_hash,
            client: escrow.client, freelancer: escrow.freelancer,
            mint: escrow.mint, gross_amount, fee_amount,
        });
        Ok(())
    }

    pub fn accept_escrow(ctx: Context<FreelancerAction>) -> Result<()> {
        transition(&mut ctx.accounts.escrow, EscrowState::Funded, EscrowState::Active)?;
        ctx.accounts.escrow.accepted_at = Clock::get()?.unix_timestamp;
        emit!(EscrowStateChanged { escrow: ctx.accounts.escrow.key(), state: EscrowState::Active });
        Ok(())
    }

    pub fn submit_work(ctx: Context<FreelancerAction>) -> Result<()> {
        transition(&mut ctx.accounts.escrow, EscrowState::Active, EscrowState::Submitted)?;
        ctx.accounts.escrow.submitted_at = Clock::get()?.unix_timestamp;
        emit!(EscrowStateChanged { escrow: ctx.accounts.escrow.key(), state: EscrowState::Submitted });
        Ok(())
    }

    pub fn approve_and_release(ctx: Context<ReleaseEscrow>) -> Result<()> {
        require_state(&ctx.accounts.escrow, EscrowState::Submitted)?;
        let surplus = ctx.accounts.vault.amount.checked_sub(ctx.accounts.escrow.gross_amount)
            .ok_or(EscrowError::VaultUnderfunded)?;
        let payout = ctx.accounts.escrow.gross_amount.checked_sub(ctx.accounts.escrow.fee_amount)
            .ok_or(EscrowError::MathOverflow)?;
        transfer_from_vault(&ctx, &ctx.accounts.freelancer_token, payout)?;
        transfer_from_vault(&ctx, &ctx.accounts.treasury_token, ctx.accounts.escrow.fee_amount)?;
        transfer_from_vault(&ctx, &ctx.accounts.client_token, surplus)?;
        let now = Clock::get()?.unix_timestamp;
        ctx.accounts.escrow.state = EscrowState::Released;
        ctx.accounts.escrow.resolved_at = now;
        emit!(EscrowSettled { escrow: ctx.accounts.escrow.key(), state: EscrowState::Released, recipient: ctx.accounts.freelancer.key(), gross_amount: ctx.accounts.escrow.gross_amount, fee_amount: ctx.accounts.escrow.fee_amount, resolved_at: now });
        close_vault(&ctx.accounts.token_program, &ctx.accounts.vault, &ctx.accounts.client.to_account_info(), &ctx.accounts.escrow)?;
        Ok(())
    }

    pub fn cancel_before_acceptance(ctx: Context<CancelEscrow>) -> Result<()> {
        require_state(&ctx.accounts.escrow, EscrowState::Funded)?;
        require!(ctx.accounts.vault.amount >= ctx.accounts.escrow.gross_amount, EscrowError::VaultUnderfunded);
        transfer_cancel_refund(&ctx, ctx.accounts.vault.amount)?;
        let now = Clock::get()?.unix_timestamp;
        ctx.accounts.escrow.state = EscrowState::Cancelled;
        ctx.accounts.escrow.resolved_at = now;
        emit!(EscrowSettled { escrow: ctx.accounts.escrow.key(), state: EscrowState::Cancelled, recipient: ctx.accounts.client.key(), gross_amount: ctx.accounts.escrow.gross_amount, fee_amount: 0, resolved_at: now });
        close_cancel_vault(&ctx)?;
        Ok(())
    }

    pub fn open_dispute(ctx: Context<ParticipantAction>) -> Result<()> {
        require!(matches!(ctx.accounts.escrow.state, EscrowState::Funded | EscrowState::Active | EscrowState::Submitted), EscrowError::InvalidState);
        ctx.accounts.escrow.state = EscrowState::Disputed;
        ctx.accounts.escrow.disputed_at = Clock::get()?.unix_timestamp;
        emit!(DisputeOpened { escrow: ctx.accounts.escrow.key(), opened_by: ctx.accounts.participant.key(), opened_at: ctx.accounts.escrow.disputed_at });
        Ok(())
    }

    pub fn resolve_dispute(ctx: Context<ResolveDispute>, decision: DisputeDecision) -> Result<()> {
        require_state(&ctx.accounts.escrow, EscrowState::Disputed)?;
        let surplus = ctx.accounts.vault.amount.checked_sub(ctx.accounts.escrow.gross_amount)
            .ok_or(EscrowError::VaultUnderfunded)?;
        let (state, recipient, fee_amount) = match decision {
            DisputeDecision::RefundClient => {
                transfer_resolution(&ctx, &ctx.accounts.client_token, ctx.accounts.vault.amount)?;
                (EscrowState::Refunded, ctx.accounts.client.key(), 0)
            }
            DisputeDecision::AwardFreelancer => {
                let payout = ctx.accounts.escrow.gross_amount.checked_sub(ctx.accounts.escrow.fee_amount)
                    .ok_or(EscrowError::MathOverflow)?;
                transfer_resolution(&ctx, &ctx.accounts.freelancer_token, payout)?;
                transfer_resolution(&ctx, &ctx.accounts.treasury_token, ctx.accounts.escrow.fee_amount)?;
                transfer_resolution(&ctx, &ctx.accounts.client_token, surplus)?;
                (EscrowState::Awarded, ctx.accounts.freelancer.key(), ctx.accounts.escrow.fee_amount)
            }
        };
        let now = Clock::get()?.unix_timestamp;
        ctx.accounts.escrow.state = state;
        ctx.accounts.escrow.resolved_at = now;
        emit!(EscrowSettled { escrow: ctx.accounts.escrow.key(), state, recipient, gross_amount: ctx.accounts.escrow.gross_amount, fee_amount, resolved_at: now });
        close_resolution_vault(&ctx)?;
        Ok(())
    }
}

#[derive(Accounts)]
pub struct InitializeConfig<'info> {
    #[account(mut)] pub admin: Signer<'info>,
    /// CHECK: Stored as the authority whose ATA receives platform fees.
    pub treasury: UncheckedAccount<'info>,
    #[account(constraint = *accepted_mint.to_account_info().owner == token_program.key() @ EscrowError::WrongTokenProgram)]
    pub accepted_mint: InterfaceAccount<'info, Mint>,
    pub token_program: Interface<'info, TokenInterface>,
    #[account(init, payer = admin, space = 8 + Config::INIT_SPACE, seeds = [b"config"], bump)]
    pub config: Box<Account<'info, Config>>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct AdminConfig<'info> {
    pub admin: Signer<'info>,
    #[account(mut, seeds = [b"config"], bump = config.bump, has_one = admin @ EscrowError::Unauthorized)]
    pub config: Account<'info, Config>,
}

#[derive(Accounts)]
pub struct AcceptAdmin<'info> {
    pub pending_admin: Signer<'info>,
    #[account(mut, seeds = [b"config"], bump = config.bump)]
    pub config: Account<'info, Config>,
}

#[derive(Accounts)]
#[instruction(contract_id_hash: [u8; 32])]
pub struct CreateAndFundEscrow<'info> {
    #[account(mut)] pub client: Signer<'info>,
    /// CHECK: Its key is committed to the escrow; it signs subsequent freelancer actions.
    pub freelancer: UncheckedAccount<'info>,
    #[account(seeds = [b"config"], bump = config.bump, has_one = accepted_mint, has_one = token_program)]
    pub config: Account<'info, Config>,
    pub accepted_mint: InterfaceAccount<'info, Mint>,
    #[account(init, payer = client, space = 8 + Escrow::INIT_SPACE, seeds = [b"escrow", config.key().as_ref(), contract_id_hash.as_ref()], bump)]
    pub escrow: Account<'info, Escrow>,
    #[account(mut, token::mint = accepted_mint, token::authority = client, token::token_program = token_program)]
    pub client_token: InterfaceAccount<'info, TokenAccount>,
    #[account(init, payer = client, associated_token::mint = accepted_mint, associated_token::authority = escrow, associated_token::token_program = token_program)]
    pub vault: InterfaceAccount<'info, TokenAccount>,
    pub token_program: Interface<'info, TokenInterface>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct FreelancerAction<'info> {
    pub freelancer: Signer<'info>,
    #[account(mut, seeds = [b"escrow", escrow.config.as_ref(), escrow.contract_id_hash.as_ref()], bump = escrow.bump, has_one = freelancer @ EscrowError::Unauthorized)]
    pub escrow: Account<'info, Escrow>,
}

#[derive(Accounts)]
pub struct ParticipantAction<'info> {
    pub participant: Signer<'info>,
    #[account(mut, seeds = [b"escrow", escrow.config.as_ref(), escrow.contract_id_hash.as_ref()], bump = escrow.bump, constraint = participant.key() == escrow.client || participant.key() == escrow.freelancer @ EscrowError::Unauthorized)]
    pub escrow: Account<'info, Escrow>,
}

#[derive(Accounts)]
pub struct ReleaseEscrow<'info> {
    #[account(mut)] pub client: Signer<'info>,
    /// CHECK: Address constrained to the escrow and used as ATA authority.
    pub freelancer: UncheckedAccount<'info>,
    /// CHECK: Address constrained to config treasury and used as ATA authority.
    pub treasury: UncheckedAccount<'info>,
    #[account(seeds = [b"config"], bump = config.bump, has_one = accepted_mint, has_one = token_program, has_one = treasury)]
    pub config: Box<Account<'info, Config>>,
    #[account(mut, close = client, seeds = [b"escrow", config.key().as_ref(), escrow.contract_id_hash.as_ref()], bump = escrow.bump, has_one = client @ EscrowError::Unauthorized, has_one = freelancer, constraint = escrow.mint == accepted_mint.key() @ EscrowError::WrongMint)]
    pub escrow: Box<Account<'info, Escrow>>,
    pub accepted_mint: Box<InterfaceAccount<'info, Mint>>,
    #[account(mut, associated_token::mint = accepted_mint, associated_token::authority = escrow, associated_token::token_program = token_program)]
    pub vault: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = client, associated_token::mint = accepted_mint, associated_token::authority = freelancer, associated_token::token_program = token_program)]
    pub freelancer_token: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = client, associated_token::mint = accepted_mint, associated_token::authority = treasury, associated_token::token_program = token_program)]
    pub treasury_token: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = client, associated_token::mint = accepted_mint, associated_token::authority = client, associated_token::token_program = token_program)]
    pub client_token: Box<InterfaceAccount<'info, TokenAccount>>,
    pub token_program: Interface<'info, TokenInterface>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct CancelEscrow<'info> {
    #[account(mut)] pub client: Signer<'info>,
    #[account(seeds = [b"config"], bump = config.bump, has_one = accepted_mint, has_one = token_program)] pub config: Account<'info, Config>,
    #[account(mut, close = client, seeds = [b"escrow", config.key().as_ref(), escrow.contract_id_hash.as_ref()], bump = escrow.bump, has_one = client @ EscrowError::Unauthorized, constraint = escrow.mint == accepted_mint.key() @ EscrowError::WrongMint)] pub escrow: Account<'info, Escrow>,
    pub accepted_mint: InterfaceAccount<'info, Mint>,
    #[account(mut, associated_token::mint = accepted_mint, associated_token::authority = escrow, associated_token::token_program = token_program)] pub vault: InterfaceAccount<'info, TokenAccount>,
    #[account(init_if_needed, payer = client, associated_token::mint = accepted_mint, associated_token::authority = client, associated_token::token_program = token_program)] pub client_token: InterfaceAccount<'info, TokenAccount>,
    pub token_program: Interface<'info, TokenInterface>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct ResolveDispute<'info> {
    #[account(mut)] pub admin: Signer<'info>,
    /// CHECK: Constrained to escrow client and receives refunds/rent.
    #[account(mut)] pub client: UncheckedAccount<'info>,
    /// CHECK: Constrained to escrow freelancer and used as ATA authority.
    pub freelancer: UncheckedAccount<'info>,
    /// CHECK: Constrained to config treasury and used as ATA authority.
    pub treasury: UncheckedAccount<'info>,
    #[account(seeds = [b"config"], bump = config.bump, has_one = admin @ EscrowError::Unauthorized, has_one = accepted_mint, has_one = token_program, has_one = treasury)] pub config: Box<Account<'info, Config>>,
    #[account(mut, close = client, seeds = [b"escrow", config.key().as_ref(), escrow.contract_id_hash.as_ref()], bump = escrow.bump, has_one = client, has_one = freelancer, constraint = escrow.mint == accepted_mint.key() @ EscrowError::WrongMint)] pub escrow: Box<Account<'info, Escrow>>,
    pub accepted_mint: Box<InterfaceAccount<'info, Mint>>,
    #[account(mut, associated_token::mint = accepted_mint, associated_token::authority = escrow, associated_token::token_program = token_program)] pub vault: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = admin, associated_token::mint = accepted_mint, associated_token::authority = client, associated_token::token_program = token_program)] pub client_token: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = admin, associated_token::mint = accepted_mint, associated_token::authority = freelancer, associated_token::token_program = token_program)] pub freelancer_token: Box<InterfaceAccount<'info, TokenAccount>>,
    #[account(init_if_needed, payer = admin, associated_token::mint = accepted_mint, associated_token::authority = treasury, associated_token::token_program = token_program)] pub treasury_token: Box<InterfaceAccount<'info, TokenAccount>>,
    pub token_program: Interface<'info, TokenInterface>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
}

#[account]
#[derive(InitSpace)]
pub struct Config {
    pub version: u8,
    pub admin: Pubkey,
    pub pending_admin: Pubkey,
    pub accepted_mint: Pubkey,
    pub token_program: Pubkey,
    pub treasury: Pubkey,
    pub fee_bps: u16,
    pub paused: bool,
    pub bump: u8,
}

#[account]
#[derive(InitSpace)]
pub struct Escrow {
    pub version: u8,
    pub config: Pubkey,
    pub contract_id_hash: [u8; 32],
    pub terms_hash: [u8; 32],
    pub client: Pubkey,
    pub freelancer: Pubkey,
    pub mint: Pubkey,
    pub gross_amount: u64,
    pub fee_amount: u64,
    pub state: EscrowState,
    pub created_at: i64,
    pub accepted_at: i64,
    pub submitted_at: i64,
    pub disputed_at: i64,
    pub resolved_at: i64,
    pub bump: u8,
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq, InitSpace)]
pub enum EscrowState { Funded, Active, Submitted, Released, Cancelled, Disputed, Refunded, Awarded }

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, PartialEq, Eq)]
pub enum DisputeDecision { RefundClient, AwardFreelancer }

#[event] pub struct ConfigInitialized { pub config: Pubkey, pub admin: Pubkey, pub mint: Pubkey, pub treasury: Pubkey, pub fee_bps: u16 }
#[event] pub struct AdminProposed { pub config: Pubkey, pub pending_admin: Pubkey }
#[event] pub struct AdminAccepted { pub config: Pubkey, pub previous_admin: Pubkey, pub admin: Pubkey }
#[event] pub struct PauseChanged { pub config: Pubkey, pub paused: bool }
#[event] pub struct EscrowFunded { pub escrow: Pubkey, pub contract_id_hash: [u8; 32], pub terms_hash: [u8; 32], pub client: Pubkey, pub freelancer: Pubkey, pub mint: Pubkey, pub gross_amount: u64, pub fee_amount: u64 }
#[event] pub struct EscrowStateChanged { pub escrow: Pubkey, pub state: EscrowState }
#[event] pub struct DisputeOpened { pub escrow: Pubkey, pub opened_by: Pubkey, pub opened_at: i64 }
#[event] pub struct EscrowSettled { pub escrow: Pubkey, pub state: EscrowState, pub recipient: Pubkey, pub gross_amount: u64, pub fee_amount: u64, pub resolved_at: i64 }

fn calculate_fee(gross: u64, fee_bps: u16) -> Result<u64> {
    gross.checked_mul(fee_bps as u64).ok_or(EscrowError::MathOverflow)?.checked_div(BPS_DENOMINATOR).ok_or(EscrowError::MathOverflow.into())
}
fn require_state(escrow: &Escrow, expected: EscrowState) -> Result<()> { require!(escrow.state == expected, EscrowError::InvalidState); Ok(()) }
fn transition(escrow: &mut Escrow, from: EscrowState, to: EscrowState) -> Result<()> { require_state(escrow, from)?; escrow.state = to; Ok(()) }

fn signer_seeds<'a>(escrow: &'a Escrow) -> [&'a [u8]; 4] { [b"escrow", escrow.config.as_ref(), escrow.contract_id_hash.as_ref(), core::slice::from_ref(&escrow.bump)] }

fn transfer_from_user<'info>(program: &Interface<'info, TokenInterface>, from: &InterfaceAccount<'info, TokenAccount>, to: &InterfaceAccount<'info, TokenAccount>, mint: &InterfaceAccount<'info, Mint>, authority: &Signer<'info>, amount: u64) -> Result<()> {
    token_interface::transfer_checked(CpiContext::new(program.to_account_info(), TransferChecked { from: from.to_account_info(), mint: mint.to_account_info(), to: to.to_account_info(), authority: authority.to_account_info() }), amount, mint.decimals)
}
fn signed_transfer<'info>(program: &Interface<'info, TokenInterface>, from: &InterfaceAccount<'info, TokenAccount>, to: &InterfaceAccount<'info, TokenAccount>, mint: &InterfaceAccount<'info, Mint>, escrow: &Account<'info, Escrow>, amount: u64) -> Result<()> {
    if amount == 0 { return Ok(()); }
    let seeds = signer_seeds(escrow); let signer = &[&seeds[..]];
    token_interface::transfer_checked(CpiContext::new_with_signer(program.to_account_info(), TransferChecked { from: from.to_account_info(), mint: mint.to_account_info(), to: to.to_account_info(), authority: escrow.to_account_info() }, signer), amount, mint.decimals)
}
fn close_vault<'info>(program: &Interface<'info, TokenInterface>, vault: &InterfaceAccount<'info, TokenAccount>, destination: &AccountInfo<'info>, escrow: &Account<'info, Escrow>) -> Result<()> {
    let seeds = signer_seeds(escrow); let signer = &[&seeds[..]];
    token_interface::close_account(CpiContext::new_with_signer(program.to_account_info(), CloseAccount { account: vault.to_account_info(), destination: destination.clone(), authority: escrow.to_account_info() }, signer))
}
fn transfer_from_vault<'info>(ctx: &Context<ReleaseEscrow<'info>>, to: &InterfaceAccount<'info, TokenAccount>, amount: u64) -> Result<()> { signed_transfer(&ctx.accounts.token_program, &ctx.accounts.vault, to, &ctx.accounts.accepted_mint, &ctx.accounts.escrow, amount) }
fn transfer_cancel_refund<'info>(ctx: &Context<CancelEscrow<'info>>, amount: u64) -> Result<()> { signed_transfer(&ctx.accounts.token_program, &ctx.accounts.vault, &ctx.accounts.client_token, &ctx.accounts.accepted_mint, &ctx.accounts.escrow, amount) }
fn close_cancel_vault<'info>(ctx: &Context<CancelEscrow<'info>>) -> Result<()> { close_vault(&ctx.accounts.token_program, &ctx.accounts.vault, &ctx.accounts.client.to_account_info(), &ctx.accounts.escrow) }
fn transfer_resolution<'info>(ctx: &Context<ResolveDispute<'info>>, to: &InterfaceAccount<'info, TokenAccount>, amount: u64) -> Result<()> { signed_transfer(&ctx.accounts.token_program, &ctx.accounts.vault, to, &ctx.accounts.accepted_mint, &ctx.accounts.escrow, amount) }
fn close_resolution_vault<'info>(ctx: &Context<ResolveDispute<'info>>) -> Result<()> { close_vault(&ctx.accounts.token_program, &ctx.accounts.vault, &ctx.accounts.client.to_account_info(), &ctx.accounts.escrow) }

#[error_code]
pub enum EscrowError {
    #[msg("Only the configured authority may perform this action")] Unauthorized,
    #[msg("The escrow is not in the required state")] InvalidState,
    #[msg("New escrow funding is paused")] Paused,
    #[msg("Amount must be greater than zero")] InvalidAmount,
    #[msg("Fee basis points must be between 0 and 10,000")] InvalidFeeBps,
    #[msg("Checked integer arithmetic failed")] MathOverflow,
    #[msg("Admin cannot be the default public key")] InvalidAdmin,
    #[msg("Client and freelancer must differ")] SameParty,
    #[msg("Terms hash cannot be zero")] InvalidTermsHash,
    #[msg("Contract identifier hash cannot be zero")] InvalidContractHash,
    #[msg("The vault contains less than the escrow's recorded gross amount")] VaultUnderfunded,
    #[msg("The escrow mint does not match the configured mint")] WrongMint,
    #[msg("The mint is not owned by the configured token program")] WrongTokenProgram,
    #[msg("Treasury cannot be the default public key")] InvalidTreasury,
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test] fn fee_is_deducted_from_gross() { assert_eq!(calculate_fee(1_000_000, 250).unwrap(), 25_000); }
    #[test] fn fee_rounds_down_in_atomic_units() { assert_eq!(calculate_fee(1, 9999).unwrap(), 0); }
    #[test] fn fee_overflow_is_rejected() { assert!(calculate_fee(u64::MAX, 10_000).is_err()); }
}
