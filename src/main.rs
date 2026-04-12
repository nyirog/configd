use std::{error::Error, future::pending};
use zbus::{connection, interface};
use zbus::zvariant::Type;
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Type)]
struct Config {
    some_int: i64,
    some_string: String,
}

struct ConfigMethod {
    config: Config,
    applied: bool,
}

#[interface(name = "org.configd.Config")]
impl ConfigMethod {
    // Can be `async` as well.
    fn get(&mut self) -> Config {
        Config {
            some_int: self.config.some_int,
            some_string: self.config.some_string.clone(),
        }
    }

    fn set(&mut self, config: Config) -> () {
        self.config = config;
        self.applied = false;
    }

    fn apply(&mut self) -> () {
        self.applied = true;
    }
}

// Although we use `tokio` here, you can use any async runtime of choice.
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let config = ConfigMethod {
        config: Config {
            some_int: 42,
            some_string: "hello".to_string(),
        },
        applied: true,
    };
    let _conn = connection::Builder::session()?
        .name("org.configd")?
        .serve_at("/org/configd", config)?
        .build()
        .await?;

    // Do other things or go to wait forever
    pending::<()>().await;

    Ok(())
}
