package kloud9.k8s.metrics

import java.sql.Struct

import com.mongodb.spark.MongoSpark
import com.typesafe.config.ConfigFactory
import org.apache.kafka.common.serialization.StringDeserializer
import org.apache.spark.internal.Logging
import org.apache.spark.sql.{SaveMode, SparkSession}
import org.apache.spark.sql.functions.{col, date_format, explode, from_json}
import org.apache.spark.sql.types._
import org.apache.spark.streaming.{Seconds, StreamingContext}
import org.apache.spark.streaming.kafka010.{ConsumerStrategies, KafkaUtils, LocationStrategies}


/* vmMetrics collects kafka data from 'VmMetrics' topic, performs spark operations and
writes the dataframe to MongoDB collection 'vmdB' in 'vm_metrics' Database
*/

object vmMetrics extends Logging{

  def main(args: Array[String]): Unit = {

    logInfo("main method started")

    /*
        Sparksession provides a single point of entry to interact
         with underlying Spark functionality and allows programming Spark with DataFrame API's
         Configurations are mentioned application.conf
       */


    val spark = SparkSession
      .builder()
      .appName(ConfigFactory.load().getString("k9.spark.appName"))
      .master(System.getenv("SPARK_MASTER"))
      .getOrCreate()

    if(spark == null){
      logError("spark session not created")
    }

    logInfo("sparkSession created")

    val sc = spark.sparkContext

    if(sc == null){
      logError("spark context not created")

    }

    logInfo("spark context created")

    //    sc.setCheckpointDir("/tmp")


    // Intialising spark streaming
    val ssc = new StreamingContext(sc, Seconds(2))


    //    if(ssc == null){
    //      logError("streaming context not created")

    //      StreamingContext.getActiveOrCreate("/tmp",() => ssc)

    //    }

    logInfo("spark streaming context created")


    var schema = new StructType()
      .add("results",ArrayType(
        new StructType()
          .add("series",ArrayType(
            new StructType()
              .add("name",StringType)
              .add("columns",ArrayType(
                StringType
              ))
             .add("values",ArrayType(ArrayType(StringType)))))))


    var vmschema = new StructType()
          .add("MetricList", ArrayType(
            new StructType()
              .add("Instance", StringType)
              .add("NodeName", StringType)
              .add("CpuAllocated", new StructType()
                .add("CpuAllocatedTimestamp", DoubleType)
                .add("CpuAllocatedValue", DoubleType))
              .add("CpuUsed", new StructType()
                .add("CpuUsedTimestamp", DoubleType)
                .add("CpuUsedValue", DoubleType))
              .add("CpuUsedMilli", new StructType()
                .add("CpuUsedMilliTimestamp", DoubleType)
                .add("CpuUsedMilliValue", DoubleType))
              .add("MemoryAllocated", new StructType()
                .add("MemAllocatedTimestamp", DoubleType)
                .add("MemAllocatedValue", DoubleType))
              .add("MemoryUsed", new StructType()
                .add("MemoryUsedTimestamp", DoubleType)
                .add("MemoryUsedValue", DoubleType))
              .add("MemoryUsedBytes", new StructType()
                .add("MemUsedByteTimestamp", DoubleType)
                .add("MemUsedByteValue", DoubleType))))


    /*
    kafkaParams - bootstrap.servers -> list of host and port pairs used for connecting to kafka brokers
       key.deserializer -> convert bytes of arrays into the desired data type
    */

    val kafkaTopics: Iterable[String] = ConfigFactory.load().getString("k9.kafka.topic") :: Nil // comma separated list of topics
    val kafkaParams = Map[String, Object]("bootstrap.servers" ->  System.getenv("KAFKA_SERVER"),
      "key.deserializer" -> classOf[StringDeserializer],
      "value.deserializer" -> classOf[StringDeserializer],
      "group.id" ->  System.getenv("GROUP_ID"))


    val stopActiveContext = true
    if (stopActiveContext) {
      StreamingContext.getActive.foreach {
        _.stop(stopSparkContext = false)
      }
    }


    val vmmetricsstream = KafkaUtils.createDirectStream(ssc, LocationStrategies.PreferConsistent, ConsumerStrategies.Subscribe[String, String](kafkaTopics, kafkaParams))


    import spark.implicits._

    vmmetricsstream.foreachRDD { rdd =>

      // converting rdd to dataframe using toDF()
      var vmmetricsstreamDf = rdd.map(record => (record.key, record.value)).toDF()

      if (!rdd.isEmpty()) {

        vmmetricsstreamDf = vmmetricsstreamDf.select($"_2" cast "string" as "json")
          .select(from_json($"json", schema) as "data")
          .select("data.*")

        var influxvmstreamDF = vmmetricsstreamDf.select($"results".getItem(0).getItem("series") as "series")
          .select($"series".getItem(0).getItem("values") as "val")
          .select($"val" (0).getItem(2) as "json") // influx timestamp will be dropped here onwards
          .select(from_json($"json", vmschema) as "List").select(explode(col("List.MetricList")) as "List") // Reads the json string into specified schema based


        influxvmstreamDF = influxvmstreamDF
          .withColumn("instance_ip",$"List.Instance")
          .withColumn("instance_name",$"List.NodeName")
          .withColumn("Timestamp",date_format(col("List.CpuAllocated.CpuAllocatedTimestamp").cast(TimestampType), "YYYY-MM-dd HH:mm:ss"))
          .withColumn("cpu_allocated_mcores",($"List.CpuAllocated.CpuAllocatedValue") * 1000)
          .withColumn("cpu_used_perc",$"List.CpuUsed.CpuUsedValue")
          .withColumn("cpu_used_mcores",$"List.CpuUsedMilli.CpuUsedMilliValue")
          .withColumn("mem_allocated_kibs",($"List.MemoryAllocated.MemAllocatedValue") / 1024)
          .withColumn("mem_used_perc",$"List.MemoryUsed.MemoryUsedValue")
          .withColumn("mem_used_kibs",($"List.MemoryUsedBytes.MemUsedByteValue") / 1024)
          .drop("List")
          .dropDuplicates()



//         checks for null

        influxvmstreamDF = influxvmstreamDF.filter(($"instance_ip" =!= "") or $"instance_ip".isNotNull)
                  .filter(($"instance_name" =!= "") or $"instance_name".isNotNull)
                  .filter(($"Timestamp" =!= "") or $"Timestamp".isNotNull)
                  .filter(($"cpu_allocated_mcores" =!= "") or $"cpu_allocated_mcores".isNotNull)
                  .filter(($"cpu_used_perc" =!= "") or $"cpu_used_perc".isNotNull)
                  .filter(($"cpu_used_mcores" =!= "") or $"cpu_used_mcores".isNotNull)
                  .filter(($"mem_allocated_kibs" =!= "") or $"mem_allocated_kibs".isNotNull)
                  .filter(($"mem_used_perc" =!= "") or $"mem_used_perc".isNotNull)
                  .filter(($"mem_used_kibs" =!= "") or $"mem_used_kibs".isNotNull)


        influxvmstreamDF.printSchema()
        influxvmstreamDF.show(false)

        logInfo("Connecting vmdB")

        //persisting the finalDF to MongoDB

        var username = System.getenv("USERNAME")
        var password = System.getenv("PASSWORD")
        var host = System.getenv("MONGOIP")


        MongoSpark.save(influxvmstreamDF.write
          .mode(SaveMode.Append)
          .option("spark.mongodb.output.collection", System.getenv("COLLECTION"))
          .option("spark.mongodb.output.uri", s"mongodb://$username:$password@$host")
          .option("spark.mongodb.output.database", System.getenv("DATABASE")))

        logInfo("Connected vmdB")

      }

    }

    ssc.start()
    ssc.awaitTermination()

    logInfo("main method ended")



  }

}

