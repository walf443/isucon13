use std::io;
use thiserror::Error;

pub mod initialize_command;
pub mod pdnsutil_command;

#[derive(Debug, Error)]
pub enum CommandError {
    #[error("IO error: {0}")]
    IoError(#[from] io::Error),
}

pub type CommandResult<T> = Result<T, CommandError>;

pub struct CommandOutput {
    pub success: bool,
    pub stdout: Vec<u8>,
    pub stderr: Vec<u8>,
}
