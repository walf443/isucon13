use async_trait::async_trait;
use isupipe_core::commands::pdnsutil_command::PDNSUtilCommand;
use isupipe_core::commands::{CommandOutput, CommandResult};

#[derive(Clone)]
pub struct PDNSUtilCommandInfra {}

#[async_trait]
impl PDNSUtilCommand for PDNSUtilCommandInfra {
    async fn add_record(
        &self,
        name: &str,
        powerdns_subdomain_address: &str,
    ) -> CommandResult<CommandOutput> {
        let output = tokio::process::Command::new("pdnsutil")
            .arg("add-record")
            .arg("u.isucon.dev")
            .arg(name)
            .arg("A")
            .arg("0")
            .arg(powerdns_subdomain_address)
            .output()
            .await?;
        Ok(CommandOutput {
            success: output.status.success(),
            stdout: output.stdout,
            stderr: output.stderr,
        })
    }
}
