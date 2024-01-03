use crate::commands::{CommandOutput, CommandResult};
use async_trait::async_trait;

#[async_trait]
pub trait InitializeCommand {
    async fn execute(&self) -> CommandResult<CommandOutput>;
}

pub trait HaveInitializeCommand {
    type Command: Sync + InitializeCommand;

    fn initialize_command(&self) -> &Self::Command;
}
