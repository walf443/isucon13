use async_trait::async_trait;
use isupipe_core::commands::initialize_command::InitializeCommand;
use isupipe_core::commands::{CommandOutput, CommandResult};

#[derive(Clone)]
pub struct InitializeCommandInfra {}

#[async_trait]
impl InitializeCommand for InitializeCommandInfra {
    async fn execute(&self) -> CommandResult<CommandOutput> {
        let output = tokio::process::Command::new("../sql/init.sh")
            .output()
            .await?;

        let result = CommandOutput {
            success: output.status.success(),
            stdout: output.stdout,
            stderr: output.stderr,
        };

        Ok(result)
    }
}
